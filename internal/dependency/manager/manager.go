package manager

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/afero"
	"golang.org/x/term"

	"github.com/g2a-com/klio/internal/context"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/dependency/registry"
	"github.com/g2a-com/klio/internal/lock"
	"github.com/g2a-com/klio/internal/log"
	"github.com/g2a-com/klio/internal/tarball"
)

const (
	indexFileName             = "dependencies.json"
	indexLockFile             = "dependencies.lock"
	dependenciesDirectoryName = "dependencies"
	defaultDirPermissions     = 0o755
)

type Updates struct {
	NonBreaking string
	Breaking    string
}

type Manager struct {
	DefaultRegistry string
	registries      map[string]registry.Registry
	os              afero.Fs
}

func NewManager() *Manager {
	return &Manager{
		registries: map[string]registry.Registry{},
		os:         afero.NewOsFs(),
	}
}

func (mgr *Manager) GetUpdateFor(dep dependency.Dependency) (Updates, error) {
	// Initialize depRegistry
	if _, ok := mgr.registries[dep.Registry]; !ok {
		mgr.registries[dep.Registry] = registry.NewLocal(dep.Registry)
		if err := mgr.registries[dep.Registry].Update(); err != nil {
			return Updates{}, err
		}
	}

	// Find versions
	depRegistry := mgr.registries[dep.Registry]

	nonBreaking, err := depRegistry.GetHighestNonBreaking(dep)
	if err != nil {
		log.Debugf("Error while checking non-breaking update: %s", err)
	}
	breaking, err := depRegistry.GetHighestBreaking(dep)
	if err != nil {
		log.Debugf("Error while checking breaking update: %s", err)
	}

	// Prepare result
	updates := Updates{}

	if nonBreaking != nil {
		updates.NonBreaking = nonBreaking.Version
	}
	if breaking != nil {
		updates.Breaking = breaking.Version
	}

	return updates, nil
}

func (mgr *Manager) InstallDependency(dep dependency.Dependency, installDir string) (*dependency.Dependency, error) {
	// == Acquire lock for updating dependencies.json ==
	// make sure main install dir exists (necessary for lockfile setup)
	if err := mgr.os.MkdirAll(installDir, defaultDirPermissions); err != nil {
		log.Fatalf("unable to create directory: %s due to %s", installDir, err)
	}
	indexLockFilePath := filepath.Join(installDir, indexLockFile)
	if err := lock.Acquire(indexLockFilePath); err != nil {
		return nil, err
	}

	// == Initialize registry ==
	dep.SetDefaults(mgr.DefaultRegistry)
	if _, ok := mgr.registries[dep.Registry]; !ok {
		if strings.HasPrefix(dep.Registry, "file:///") {
			mgr.registries[dep.Registry] = registry.NewLocal(dep.Registry)
		} else {
			mgr.registries[dep.Registry] = registry.NewRemote(dep.Registry)
		}
		if err := mgr.registries[dep.Registry].Update(); err != nil {
			return nil, err
		}
	}

	// == Search for a suitable version ==
	registryEntry, _ := mgr.registries[dep.Registry].GetExactMatch(dep)
	if registryEntry == nil {
		return nil, fmt.Errorf("cannot find %s@%s in %s", dep.Name, dep.Version, dep.Registry)
	}

	// == Download tarball to a temporary file ==
	tempFile, err := afero.TempFile(mgr.os, "", "klio-")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = mgr.os.Remove(tempFile.Name())
	}()
	checksum, err := downloadFile(registryEntry.URL, tempFile)
	if err != nil {
		return nil, err
	}

	// == Verify checksum ==
	if registryEntry.Checksum != "" && registryEntry.Checksum != checksum {
		return nil, fmt.Errorf(`checksum of the archive (%s) is different from the one specified in the regsitry (%s)`, checksum, registryEntry.Checksum)
	}

	// == Prepare directory to install the dependency ==
	outputRelPath := filepath.Join(dependenciesDirectoryName, checksum)
	outputAbsPath := filepath.Join(installDir, outputRelPath)
	if err := mgr.os.RemoveAll(outputAbsPath); err != nil {
		log.Fatalf("unable to remove directory: %s due to %s", outputAbsPath, err)
	}
	if err := mgr.os.MkdirAll(outputAbsPath, defaultDirPermissions); err != nil {
		log.Fatalf("unable to create directory: %s due to %s", outputAbsPath, err)
	}

	// == Extract tarball into the installation directory ==
	_, _ = tempFile.Seek(0, io.SeekStart)
	if err := tarball.Extract(tempFile, outputAbsPath); err != nil {
		return nil, err
	}

	// == Add dependency to dependencies.json ==
	indexFilePath := filepath.Join(installDir, indexFileName)
	index, err := dependency.LoadDependenciesIndex(indexFilePath)
	if err != nil {
		return nil, err
	}
	var newEntries []dependency.DependenciesIndexEntry
	for _, entry := range index.Entries {
		if entry.Alias != dep.Alias {
			newEntries = append(newEntries, entry)
		}
	}
	index.Entries = append(newEntries, dependency.DependenciesIndexEntry{
		Alias:    dep.Alias,
		Registry: dep.Registry,
		Name:     dep.Name,
		Version:  registryEntry.Version,
		OS:       registryEntry.OS,
		Arch:     registryEntry.Arch,
		Checksum: registryEntry.Checksum,
		Path:     outputRelPath,
	})
	if err := dependency.SaveDependenciesIndex(index); err != nil {
		return nil, err
	}

	// Release lock
	_ = lock.Release(indexLockFilePath)

	// Return info about installed dependency
	result := dep // TODO: WHY?!
	result.Version = registryEntry.Version

	return &result, nil
}

func GetInstalledCommands(paths context.Paths) []dependency.DependenciesIndexEntry {
	var entries []dependency.DependenciesIndexEntry
	discoveredPaths := []string{paths.ProjectInstallDir, paths.GlobalInstallDir}

	for _, path := range discoveredPaths {
		if path == "" {
			continue
		}
		indexPath := filepath.Join(path, indexFileName)
		indexData, err := dependency.LoadDependenciesIndex(indexPath)
		if err != nil {
			log.Debugf("can't load dependency indices from %s: %s", path, err)
			continue
		}

		// dependencies.json contains relative paths for commands, make them absolute
		for _, entry := range indexData.Entries {
			entry.Path = filepath.Join(path, entry.Path)
			entries = append(entries, entry)
		}
	}

	return entries
}

func downloadFile(url string, file io.Writer) (checksum string, err error) {
	log.Verbosef("Downloading %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	hash := sha256.New()
	writer := io.MultiWriter(file, hash)

	if term.IsTerminal(int(os.Stdout.Fd())) {
		progress := progressbar.DefaultBytes(
			resp.ContentLength, // value -1 indicates that the length is unknown
			"Downloading",
		)
		writer = io.MultiWriter(writer, progress)
	}

	if _, err = io.Copy(writer, resp.Body); err != nil {
		return "", err
	}

	return fmt.Sprintf("sha256-%x", hash.Sum(nil)), nil
}
