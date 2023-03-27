package manager

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/g2a-com/klio/internal/context"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/dependency/registry"
	"github.com/g2a-com/klio/internal/lock"
	"github.com/g2a-com/klio/internal/log"
	"github.com/g2a-com/klio/internal/tarball"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/afero"
	"golang.org/x/term"
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
	DefaultRegistry        string
	registries             map[string]registry.Registry
	os                     afero.Fs
	httpDownloadClient     *http.Client
	dependencyIndexHandler dependency.IndexHandler
	createLock             func(string) (lock.Lock, error)
}

// NewManager returns a new default Manager.
func NewManager() *Manager {
	return &Manager{
		registries:             map[string]registry.Registry{},
		os:                     afero.NewOsFs(),
		dependencyIndexHandler: &dependency.LocalIndexHandler{},
		httpDownloadClient:     http.DefaultClient,
		createLock:             lock.New,
	}
}

// GetUpdateFor gets updates for given dependency dep.
// If error doesn't occur, both major and minor updates are returned.
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

// InstallDependency installs a single dependency in the installDir directory.
// Dependency metadata is provided in dep.
func (mgr *Manager) InstallDependency(dep *dependency.Dependency, installDir string) error {
	// == Acquire lock for updating dependencies.json ==
	// make sure main install dir exists (necessary for lockfile setup)
	if err := mgr.os.MkdirAll(installDir, defaultDirPermissions); err != nil {
		log.Fatalf("unable to create directory: %s due to %s", installDir, err)
	}
	installLock, err := mgr.createLock(filepath.Join(installDir, indexLockFile))
	if err != nil {
		return err
	}
	if err := installLock.Acquire(); err != nil {
		return err
	}
	defer func() {
		err = installLock.Release()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// == Initialize registry ==
	dep.SetDefaults(mgr.DefaultRegistry)
	if _, ok := mgr.registries[dep.Registry]; !ok {
		if strings.HasPrefix(dep.Registry, "file:///") {
			mgr.registries[dep.Registry] = registry.NewLocal(dep.Registry)
		} else {
			mgr.registries[dep.Registry] = registry.NewRemote(dep.Registry)
		}
		if err := mgr.registries[dep.Registry].Update(); err != nil {
			return err
		}
	}

	// == Search for a suitable version ==
	registryEntry, _ := mgr.registries[dep.Registry].GetExactMatch(*dep)
	if registryEntry == nil {
		return &CantFindExactVersionMatchError{dep.Name, dep.Version, dep.Registry}
	}

	// == Download tarball to a temporary file ==
	tempFile, err := afero.TempFile(mgr.os, "", "klio-")
	if err != nil {
		return err
	}
	defer func() {
		_ = mgr.os.Remove(tempFile.Name())
	}()
	checksum, err := downloadFile(mgr.httpDownloadClient, registryEntry.URL, tempFile)
	if err != nil {
		return err
	}

	// == Verify checksum ==
	if registryEntry.Checksum != "" && registryEntry.Checksum != checksum {
		return fmt.Errorf(`checksum of the archive (%s) is different from the one specified in the regsitry (%s)`, checksum, registryEntry.Checksum)
	}

	// == Prepare directory to install the dependency ==
	outputRelPath := filepath.Join(dependenciesDirectoryName, checksum)
	outputAbsPath := filepath.Join(installDir, outputRelPath)
	if err := mgr.os.RemoveAll(outputAbsPath); err != nil {
		return fmt.Errorf("unable to remove directory: %s due to %s", outputAbsPath, err)
	}
	if err := mgr.os.MkdirAll(outputAbsPath, defaultDirPermissions); err != nil {
		return fmt.Errorf("unable to create directory: %s due to %s", outputAbsPath, err)
	}

	// == Extract tarball into the installation directory ==
	_, _ = tempFile.Seek(0, io.SeekStart)
	if err := tarball.Extract(tempFile, mgr.os, outputAbsPath); err != nil {
		return err
	}

	// == Add dependency to dependencies.json ==
	indexFilePath := filepath.Join(installDir, indexFileName)
	err = mgr.dependencyIndexHandler.LoadDependencyIndex(indexFilePath)
	if err != nil {
		return err
	}
	var newEntries, oldEntries []dependency.DependenciesIndexEntry
	for _, entry := range mgr.dependencyIndexHandler.GetEntries() {
		if entry.Alias != dep.Alias {
			newEntries = append(newEntries, entry)
		} else if entry.Checksum != checksum {
			oldEntries = append(oldEntries, entry)
		}
	}

	// == Remove directories with dependencies that won't be used anymore ==
	for _, entry := range oldEntries {
		absPath := filepath.Join(installDir, entry.Path)
		if err := mgr.os.RemoveAll(absPath); err != nil {
			return fmt.Errorf("unable to remove directory: %s due to %s", absPath, err)
		}
	}

	dep.Version = registryEntry.Version

	mgr.dependencyIndexHandler.SetEntries(
		append(newEntries, dependency.DependenciesIndexEntry{
			Alias:    dep.Alias,
			Registry: dep.Registry,
			Name:     dep.Name,
			Version:  registryEntry.Version,
			OS:       registryEntry.OS,
			Arch:     registryEntry.Arch,
			Checksum: registryEntry.Checksum,
			Path:     outputRelPath,
		}),
	)
	if err := mgr.dependencyIndexHandler.SaveDependencyIndex(); err != nil {
		return err
	}

	return nil
}

// RemoveDependency removes a single dependency in the installDir directory.
// Dependency metadata is provided in dep.
func (mgr *Manager) RemoveDependency(dep *dependency.Dependency, installDir string) error {
	// == Acquire lock for updating dependencies.json ==
	// make sure main install dir exists (necessary for lockfile setup)
	installLock, err := mgr.createLock(filepath.Join(installDir, indexLockFile))
	if err != nil {
		return err
	}
	if err := installLock.Acquire(); err != nil {
		return err
	}
	defer func() { _ = installLock.Release() }()

	// == Load dependencies.json ==
	indexFilePath := filepath.Join(installDir, indexFileName)
	err = mgr.dependencyIndexHandler.LoadDependencyIndex(indexFilePath)
	if err != nil {
		return err
	}

	// == Gather entires that should be removed ==
	var entriesToRemove []dependency.DependenciesIndexEntry
	for _, entry := range mgr.dependencyIndexHandler.GetEntries() {
		if entry.Alias == dep.Alias {
			entriesToRemove = append(entriesToRemove, entry)
		}
	}

	// == Assume command was deleted manually and exit ==
	if len(entriesToRemove) == 0 {
		log.Debugf("skipping removal of %s as it does not exist", dep.Alias)
		return nil
	}

	// == Remove command from filesystem if it exists ==
	for _, entry := range entriesToRemove {
		absPath := filepath.Join(installDir, entry.Path)
		if err := mgr.os.RemoveAll(absPath); err != nil {
			return fmt.Errorf("unable to delete directory: %s due to %s", absPath, err)
		}
	}

	// == Update dependencies.json ==
	mgr.dependencyIndexHandler.SetEntries(
		removeFromDependencyIndexList(
			entriesToRemove,
			mgr.dependencyIndexHandler.GetEntries(),
		))
	if err := mgr.dependencyIndexHandler.SaveDependencyIndex(); err != nil {
		return err
	}

	return nil
}

// GetInstalledCommands returns all the dependencies that are installed locally (both globally and within project scope).
// paths are the source of global and project install directories.
func (mgr *Manager) GetInstalledCommands(paths context.Paths) []dependency.DependenciesIndexEntry {
	var entries []dependency.DependenciesIndexEntry
	discoveredPaths := []string{paths.ProjectInstallDir, paths.GlobalInstallDir}

	for _, path := range discoveredPaths {
		if path == "" {
			continue
		}
		indexPath := filepath.Join(path, indexFileName)
		err := mgr.dependencyIndexHandler.LoadDependencyIndex(indexPath)
		if err != nil {
			log.Debugf("can't load dependency indices from %s: %s", path, err)
			continue
		}

		// dependencies.json contains relative paths for commands, make them absolute
		for _, entry := range mgr.dependencyIndexHandler.GetEntries() {
			entry.Path = filepath.Join(path, entry.Path)
			entries = append(entries, entry)
		}
	}

	return entries
}

func downloadFile(artifactoryClient *http.Client, url string, file io.Writer) (checksum string, err error) {
	log.Verbosef("Downloading %s", url)

	resp, err := artifactoryClient.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("artifactory responded: %s", resp.Status)
	}

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

func removeFromDependencyIndexList(entriesToRemove []dependency.DependenciesIndexEntry, entryList []dependency.DependenciesIndexEntry) []dependency.DependenciesIndexEntry {
	newEntryList := make([]dependency.DependenciesIndexEntry, 0)
	entryMap := make(map[string]struct{})
	for _, entry := range entriesToRemove {
		entryMap[entry.Alias] = struct{}{}
	}

	for _, entry := range entryList {
		if _, found := entryMap[entry.Alias]; !found {
			newEntryList = append(newEntryList, entry)
		}
	}
	return newEntryList
}
