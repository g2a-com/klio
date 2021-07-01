package manager

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/g2a-com/klio/internal/context"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/dependency/registry"
	"github.com/g2a-com/klio/internal/lock"
	"github.com/g2a-com/klio/internal/log"
	"github.com/g2a-com/klio/internal/tarball"
	"github.com/schollz/progressbar/v3"
)

type ScopeType string

const (
	GlobalScope  ScopeType = "global"
	ProjectScope ScopeType = "project"
)

type Updates struct {
	NonBreaking string
	Breaking    string
}

type Manager struct {
	DefaultRegistry string
	registries      map[string]*registry.Registry
	context         context.CLIContext
}

func NewManager(ctx context.CLIContext) *Manager {
	return &Manager{
		context: ctx,
	}
}

func (mgr *Manager) GetUpdateFor(dep dependency.Dependency) (Updates, error) {
	// Initialize registry
	if _, ok := mgr.registries[dep.Registry]; !ok {
		if mgr.registries == nil {
			mgr.registries = map[string]*registry.Registry{}
		}
		mgr.registries[dep.Registry] = registry.New(dep.Registry)
		if err := mgr.registries[dep.Registry].Update(); err != nil {
			return Updates{}, err
		}
	}

	// Find versions
	reg := mgr.registries[dep.Registry]

	nonBreaking := reg.FindCompatibleDependency(dependency.Dependency{
		Registry: dep.Registry,
		Name:     dep.Name,
		Version:  fmt.Sprintf("> %s, ^ %s", dep.Version, dep.Version),
	})
	breaking := reg.FindCompatibleDependency(dependency.Dependency{
		Registry: dep.Registry,
		Name:     dep.Name,
		Version:  "> " + dep.Version,
	})

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

func (mgr *Manager) InstallDependency(dep dependency.Dependency, scope ScopeType) (*dependency.Dependency, error) {
	installDir, err := mgr.getInstallDir(scope)
	if err != nil {
		return nil, err
	}
	indexPath := filepath.Join(installDir, "dependencies.json")
	indexLockfilePath := filepath.Join(installDir, "dependencies.lock")

	// Fill missing values
	if dep.Alias == "" {
		dep.Alias = dep.Name
	}
	if dep.Registry == "" {
		dep.Registry = mgr.DefaultRegistry
	}

	// Make sure that install directory exists
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		if err := os.MkdirAll(installDir, 0755); err != nil {
			log.LogfAndExit(log.FatalLevel, "unable to create directory: %s due to %s", installDir, err)
		}
	}

	// Acquire lock for updating dependencies.json
	if err := lock.Acquire(indexLockfilePath); err != nil {
		return nil, err
	}

	// Initialize registry
	if _, ok := mgr.registries[dep.Registry]; !ok {
		if mgr.registries == nil {
			mgr.registries = map[string]*registry.Registry{}
		}
		mgr.registries[dep.Registry] = registry.New(dep.Registry)
		if err := mgr.registries[dep.Registry].Update(); err != nil {
			return nil, err
		}
	}

	// Search for a suitable version
	entry := mgr.registries[dep.Registry].FindCompatibleDependency(dep)
	if entry == nil {
		return nil, fmt.Errorf("cannot find %s@%s in %s", dep.Name, dep.Version, dep.Registry)
	}

	// Create temporary file for a tarball
	file, err := ioutil.TempFile("", "klio-")
	if err != nil {
		return nil, err
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(file.Name())

	// Download tarball
	checksum, err := downloadFile(entry.URL, file)
	if err != nil {
		return nil, err
	}

	// Verify checksum
	if entry.Checksum != "" && entry.Checksum != checksum {
		return nil, fmt.Errorf(`checksum of the archive (%s) is different from the one specified in the regsitry (%s)`, checksum, entry.Checksum)
	}
	if dep.Checksum != "" && dep.Checksum != checksum {
		return nil, fmt.Errorf(`checksum of the archive (%s) is different than expected (%s)`, checksum, dep.Checksum)
	}

	// Make sure that output dir exists and is empty
	outputRelPath := filepath.Join("dependencies", checksum)
	outputAbsPath := filepath.Join(installDir, outputRelPath)
	if _, err := os.Stat(outputAbsPath); err == nil {
		_ = os.RemoveAll(outputAbsPath)
	} else if !os.IsNotExist(err) {
		log.LogfAndExit(log.FatalLevel, "unable to remove directory: %s due to %s", outputAbsPath, err)
	}
	if err := os.MkdirAll(outputAbsPath, 0755); err != nil {
		log.LogfAndExit(log.FatalLevel, "unable to create directory: %s due to %s", outputAbsPath, err)
	}

	// Extract tarball
	_, _ = file.Seek(0, io.SeekStart)
	if err := tarball.Extract(file, outputAbsPath); err != nil {
		return nil, err
	}

	// Add dependency to dependencies.json
	index, err := dependency.LoadIndex(indexPath)
	if err != nil {
		return nil, err
	}
	newEntries := make([]dependency.DependenciesIndexEntry, 0, len(index.Entries))
	for _, entry := range index.Entries {
		if entry.Alias != dep.Alias {
			newEntries = append(newEntries, entry)
		}
	}
	newEntries = append(newEntries, dependency.DependenciesIndexEntry{
		Alias:    dep.Alias,
		Registry: dep.Registry,
		Name:     dep.Name,
		Version:  entry.Version,
		OS:       entry.OS,
		Arch:     entry.Arch,
		Checksum: entry.Checksum,
		Path:     outputRelPath,
	})
	index.Entries = newEntries
	if err := dependency.SaveIndex(index); err != nil {
		return nil, err
	}

	// Release lock
	_ = lock.Release(indexLockfilePath)

	// Return info about installed dependency
	result := dep
	result.Version = entry.Version
	result.Checksum = checksum

	return &result, nil
}

func (mgr *Manager) GetInstalledCommands(scope ScopeType) ([]dependency.DependenciesIndexEntry, error) {
	installDir, err := mgr.getInstallDir(scope)
	if err != nil {
		return []dependency.DependenciesIndexEntry{}, err
	}

	indexPath := filepath.Join(installDir, "dependencies.json")
	indexData, err := dependency.LoadIndex(indexPath)
	if err != nil {
		return []dependency.DependenciesIndexEntry{}, err
	}

	// dependencies.json contains relative paths for commands, make them absolute
	for idx := range indexData.Entries {
		indexData.Entries[idx].Path = filepath.Join(installDir, indexData.Entries[idx].Path)
	}

	return indexData.Entries, nil
}

func (mgr *Manager) getInstallDir(scope ScopeType) (string, error) {
	switch scope {
	case GlobalScope:
		if mgr.context.Paths.GlobalInstallDir == "" {
			return "", errors.New("cannot find global directory")
		}
		return mgr.context.Paths.GlobalInstallDir, nil
	case ProjectScope:
		if mgr.context.Paths.ProjectInstallDir == "" {
			return "", errors.New("cannot find project directory")
		}
		return mgr.context.Paths.ProjectInstallDir, nil
	default:
		return "", fmt.Errorf("unknown scope: %s", scope)
	}
}

func downloadFile(url string, file io.Writer) (checksum string, err error) {
	log.Verbosef("Downloading %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	hash := sha256.New()
	writer := io.MultiWriter(file, hash)

	if terminal.IsTerminal(int(os.Stdout.Fd())) {
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
