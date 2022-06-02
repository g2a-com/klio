package manager

import (
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/g2a-com/klio/internal/dependency/registry"
	"github.com/g2a-com/klio/internal/lock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/g2a-com/klio/internal/context"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/stretchr/testify/suite"
)

const (
	validProjectInstallPath   = "valid/project/install/path"
	invalidProjectInstallPath = "some/messed/up/project/path"

	dependencyName = "dosomething"
)

var localTestDependencyIndexEntries = []string{"first/entry", "second/entry"}

// ============================================
// =========== TEST SUITE START ===============
// ============================================

type managerTestingSuite struct {
	suite.Suite
	// Dependency Manager
	mgr *Manager

	// Mock http remote registry
	RemoteHttpRegistryClient *http.Client
	// List of all mock registries with predefined mock http servers
	MockRegistries map[string]registry.Registry
	// Configuration paths for Klio
	Paths context.Paths
	// Default registry to query for dependency
	DefaultRegistry string
	// Dependency to get update for
	DepToUpdate dependency.Dependency
	// Dependency to install
	DepToInstall dependency.Dependency
	// Simulates problems with http server that provides command registry
	IsDependencyServerFaulty bool
	// Place to install your dependency
	InstallDir string
	// Function providing install lock
	InstallLock func(string) (lock.Lock, error)

	// Desired list of versions available locally
	ExpectedDepsInstalledLocally []dependency.DependenciesIndexEntry
	// Desired list of versions available
	ExpectedUpdates Updates
	// Desired dependency structure that was mock installed
	ExpectedInstalledDependency dependency.Dependency
	// Flag marking that there is an error while checking for dependencies
	CheckForUpdatesShouldFailWith error
	// Flag marking that there is an error while installing a dependency
	CommandInstallShouldFailWith error
}

type IndexHandlerError struct{}

func (e *IndexHandlerError) Error() string { return "acquire lock error" }

func (s *managerTestingSuite) SetupTest() {
	var localTestDependencyIndex []dependency.DependenciesIndexEntry
	for _, e := range localTestDependencyIndexEntries {
		localTestDependencyIndex = append(localTestDependencyIndex, dependency.DependenciesIndexEntry{
			Path: e,
		})
	}

	indexHandler := new(mockIndexHandler)
	indexHandler.On("LoadDependencyIndex", filepath.Join(validProjectInstallPath, "dependencies.json")).Return(nil)
	indexHandler.On("LoadDependencyIndex", filepath.Join(invalidProjectInstallPath, "dependencies.json")).Return(&IndexHandlerError{})
	indexHandler.On("GetEntries").Return(localTestDependencyIndex)
	indexHandler.On("SetEntries", mock.Anything)
	indexHandler.On("SaveDependencyIndex").Return(nil)

	s.mgr = &Manager{
		DefaultRegistry:        s.DefaultRegistry,
		registries:             s.MockRegistries,
		os:                     getMockFs(),
		httpDownloadClient:     s.RemoteHttpRegistryClient,
		dependencyIndexHandler: indexHandler,
		createLock:             s.InstallLock,
	}
}

func (s *managerTestingSuite) TestGetInstalledCommands() {
	listOfLocallyInstalledDeps := s.mgr.GetInstalledCommands(s.Paths)
	assert.Equal(s.T(), s.ExpectedDepsInstalledLocally, listOfLocallyInstalledDeps)
}

func (s *managerTestingSuite) TestGetUpdateFor() {
	availableUpdates, err := s.mgr.GetUpdateFor(s.DepToUpdate)
	if s.CheckForUpdatesShouldFailWith != nil {
		assert.ErrorContains(s.T(), err, s.CheckForUpdatesShouldFailWith.Error())
	} else {
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), s.ExpectedUpdates, availableUpdates)
	}
}

func (s *managerTestingSuite) TestManagerInstallDependency() {
	depToInstall := s.DepToInstall
	err := s.mgr.InstallDependency(&depToInstall, s.InstallDir)
	if s.CommandInstallShouldFailWith != nil {
		assert.ErrorContains(s.T(), err, s.CommandInstallShouldFailWith.Error())
	} else {
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), s.ExpectedInstalledDependency, depToInstall)
	}
}

// ============================================
// ============ TEST SUITE END ================
// ============================================

func TestInstallMinorUpgradeWithIndexHandlerMalfunction(t *testing.T) {
	remoteHttpRegistry := httptest.NewServer(&testHandler{})
	defer remoteHttpRegistry.Close()

	depToUpdate := dependency.Dependency{Name: dependencyName, Registry: remoteHttpRegistry.URL, Alias: dependencyName, Version: "2.12.0"}
	depToInstall := dependency.Dependency{Name: dependencyName, Registry: remoteHttpRegistry.URL, Alias: dependencyName, Version: "2.12.1"}
	singleEntry := registry.Entry{
		Name:    dependencyName,
		Version: "2.12.1",
		OS:      "windows",
		Arch:    "amd64",
		URL:     fmt.Sprintf("%s/%s/%s.tar.gz", remoteHttpRegistry.URL, "registry/commands", dependencyName),
	}

	r := new(mockRegistry)
	r.On("GetHighestBreaking", depToUpdate).Return(&registry.Entry{}, fmt.Errorf("no breaking change avaialable"))
	r.On("GetHighestNonBreaking", depToUpdate).Return(&singleEntry, nil)
	r.On("GetExactMatch", depToInstall).Return(&singleEntry, nil)
	allRegistries := map[string]registry.Registry{
		remoteHttpRegistry.URL: r,
	}

	var expectedListOfInstalledCommands []dependency.DependenciesIndexEntry

	mts := managerTestingSuite{
		Paths:                    context.Paths{ProjectInstallDir: invalidProjectInstallPath},
		DefaultRegistry:          remoteHttpRegistry.URL,
		RemoteHttpRegistryClient: remoteHttpRegistry.Client(),
		MockRegistries:           allRegistries,
		DepToUpdate:              depToUpdate,
		DepToInstall:             depToInstall,
		IsDependencyServerFaulty: false,
		InstallDir:               invalidProjectInstallPath,
		InstallLock:              newMockLock,

		ExpectedDepsInstalledLocally: expectedListOfInstalledCommands,
		ExpectedUpdates:              Updates{NonBreaking: singleEntry.Version, Breaking: ""},
		ExpectedInstalledDependency: dependency.Dependency{
			Name:     dependencyName,
			Registry: remoteHttpRegistry.URL,
			Version:  "2.12.1",
			Alias:    dependencyName,
		},
		CommandInstallShouldFailWith: &IndexHandlerError{},
	}

	suite.Run(t, &mts)
}

func TestInstallMinorUpgradeWithAcquireLockFailing(t *testing.T) {
	remoteHttpRegistry := httptest.NewServer(&testHandler{})
	defer remoteHttpRegistry.Close()

	depToUpdate := dependency.Dependency{Name: dependencyName, Registry: remoteHttpRegistry.URL, Alias: dependencyName, Version: "2.12.0"}
	depToInstall := dependency.Dependency{Name: dependencyName, Registry: remoteHttpRegistry.URL, Alias: dependencyName, Version: "2.12.1"}
	singleEntry := registry.Entry{
		Name:    dependencyName,
		Version: "2.12.1",
		OS:      "windows",
		Arch:    "amd64",
		URL:     fmt.Sprintf("%s/%s/%s.tar.gz", remoteHttpRegistry.URL, "registry/commands", dependencyName),
	}

	r := new(mockRegistry)
	r.On("GetHighestBreaking", depToUpdate).Return(&registry.Entry{}, fmt.Errorf("no breaking change avaialable"))
	r.On("GetHighestNonBreaking", depToUpdate).Return(&singleEntry, nil)
	r.On("GetExactMatch", depToInstall).Return(&singleEntry, nil)
	allRegistries := map[string]registry.Registry{
		remoteHttpRegistry.URL: r,
	}

	var expectedListOfInstalledCommands []dependency.DependenciesIndexEntry
	for _, e := range localTestDependencyIndexEntries {
		expectedListOfInstalledCommands = append(expectedListOfInstalledCommands, dependency.DependenciesIndexEntry{
			Path: filepath.Join(validProjectInstallPath, e),
		})
	}

	mts := managerTestingSuite{
		Paths:                    context.Paths{ProjectInstallDir: validProjectInstallPath},
		DefaultRegistry:          remoteHttpRegistry.URL,
		RemoteHttpRegistryClient: remoteHttpRegistry.Client(),
		MockRegistries:           allRegistries,
		DepToUpdate:              depToUpdate,
		DepToInstall:             depToInstall,
		IsDependencyServerFaulty: false,
		InstallDir:               validProjectInstallPath,
		InstallLock:              newMockLockFailingToAcquire,

		ExpectedDepsInstalledLocally: expectedListOfInstalledCommands,
		ExpectedUpdates:              Updates{NonBreaking: singleEntry.Version, Breaking: ""},
		ExpectedInstalledDependency: dependency.Dependency{
			Name:     dependencyName,
			Registry: remoteHttpRegistry.URL,
			Version:  "2.12.1",
			Alias:    dependencyName,
		},
		CommandInstallShouldFailWith: &AcquireLockError{},
	}

	suite.Run(t, &mts)
}

func TestInstallMinorUpgradeInCanonScenario(t *testing.T) {
	remoteHttpRegistry := httptest.NewServer(&testHandler{})
	defer remoteHttpRegistry.Close()

	depToUpdate := dependency.Dependency{Name: dependencyName, Registry: remoteHttpRegistry.URL, Alias: dependencyName, Version: "2.12.0"}
	depToInstall := dependency.Dependency{Name: dependencyName, Registry: remoteHttpRegistry.URL, Alias: dependencyName, Version: "2.12.1"}
	singleEntry := registry.Entry{
		Name:    dependencyName,
		Version: "2.12.1",
		OS:      "windows",
		Arch:    "amd64",
		URL:     fmt.Sprintf("%s/%s/%s.tar.gz", remoteHttpRegistry.URL, "registry/commands", dependencyName),
	}

	r := new(mockRegistry)
	r.On("GetHighestBreaking", depToUpdate).Return(&registry.Entry{}, fmt.Errorf("no breaking change avaialable"))
	r.On("GetHighestNonBreaking", depToUpdate).Return(&singleEntry, nil)
	r.On("GetExactMatch", depToInstall).Return(&singleEntry, nil)
	allRegistries := map[string]registry.Registry{
		remoteHttpRegistry.URL: r,
	}

	var expectedListOfInstalledCommands []dependency.DependenciesIndexEntry
	for _, e := range localTestDependencyIndexEntries {
		expectedListOfInstalledCommands = append(expectedListOfInstalledCommands, dependency.DependenciesIndexEntry{
			Path: filepath.Join(validProjectInstallPath, e),
		})
	}

	mts := managerTestingSuite{
		Paths:                    context.Paths{ProjectInstallDir: validProjectInstallPath},
		DefaultRegistry:          remoteHttpRegistry.URL,
		RemoteHttpRegistryClient: remoteHttpRegistry.Client(),
		MockRegistries:           allRegistries,
		DepToUpdate:              depToUpdate,
		DepToInstall:             depToInstall,
		IsDependencyServerFaulty: false,
		InstallDir:               validProjectInstallPath,
		InstallLock:              newMockLock,

		ExpectedDepsInstalledLocally: expectedListOfInstalledCommands,
		ExpectedUpdates:              Updates{NonBreaking: singleEntry.Version, Breaking: ""},
		ExpectedInstalledDependency: dependency.Dependency{
			Name:     dependencyName,
			Registry: remoteHttpRegistry.URL,
			Version:  "2.12.1",
			Alias:    dependencyName,
		},
	}

	suite.Run(t, &mts)
}

func TestInstallMajorUpdateWithDefaultRegistryFallbackOnDependency(t *testing.T) {
	remoteHttpRegistry := httptest.NewServer(&testHandler{})
	defer remoteHttpRegistry.Close()

	depToUpdate := dependency.Dependency{Name: dependencyName, Registry: remoteHttpRegistry.URL, Alias: dependencyName, Version: "2.11.2"}
	depToInstall := dependency.Dependency{Name: dependencyName, Registry: remoteHttpRegistry.URL, Alias: dependencyName, Version: "2.12.1"}
	singleEntry := registry.Entry{
		Name:    dependencyName,
		Version: "2.12.1",
		OS:      "windows",
		Arch:    "amd64",
		URL:     fmt.Sprintf("%s/%s/%s.tar.gz", remoteHttpRegistry.URL, "registry/commands", dependencyName),
	}

	r := new(mockRegistry)
	r.On("GetHighestNonBreaking", depToUpdate).Return(&registry.Entry{}, fmt.Errorf("no non-breaking change avaialable"))
	r.On("GetHighestBreaking", depToUpdate).Return(&singleEntry, nil)
	r.On("GetExactMatch", depToInstall).Return(&singleEntry, nil)
	allRegistries := map[string]registry.Registry{
		remoteHttpRegistry.URL: r,
	}

	var expectedListOfInstalledCommands []dependency.DependenciesIndexEntry
	for _, e := range localTestDependencyIndexEntries {
		expectedListOfInstalledCommands = append(expectedListOfInstalledCommands, dependency.DependenciesIndexEntry{
			Path: filepath.Join(validProjectInstallPath, e),
		})
	}

	mts := managerTestingSuite{
		Paths:                    context.Paths{ProjectInstallDir: validProjectInstallPath},
		DefaultRegistry:          "",
		RemoteHttpRegistryClient: remoteHttpRegistry.Client(),
		MockRegistries:           allRegistries,
		DepToUpdate:              depToUpdate,
		DepToInstall:             depToInstall,
		IsDependencyServerFaulty: false,
		InstallDir:               validProjectInstallPath,
		InstallLock:              newMockLock,

		ExpectedDepsInstalledLocally: expectedListOfInstalledCommands,
		ExpectedUpdates:              Updates{Breaking: singleEntry.Version, NonBreaking: ""},
		ExpectedInstalledDependency: dependency.Dependency{
			Name:     dependencyName,
			Registry: remoteHttpRegistry.URL,
			Version:  "2.12.1",
			Alias:    dependencyName,
		},
	}

	suite.Run(t, &mts)
}

func TestInstallMajorUpdateWithoutDefaultRegistryFallback(t *testing.T) {
	remoteHttpRegistry := httptest.NewServer(&testHandler{})
	defer remoteHttpRegistry.Close()

	depToUpdate := dependency.Dependency{Name: dependencyName, Registry: "", Alias: dependencyName, Version: "2.11.2"}
	depToInstall := dependency.Dependency{Name: dependencyName, Registry: "", Alias: dependencyName, Version: "2.12.1"}
	singleEntry := registry.Entry{
		Name:    dependencyName,
		Version: "2.12.1",
		OS:      "windows",
		Arch:    "amd64",
		URL:     fmt.Sprintf("%s/%s/%s.tar.gz", remoteHttpRegistry.URL, "registry/commands", dependencyName),
	}

	r := new(mockRegistry)
	r.On("GetHighestNonBreaking", depToUpdate).Return(&registry.Entry{}, fmt.Errorf("no non-breaking change avaialable"))
	r.On("GetHighestBreaking", depToUpdate).Return(&singleEntry, nil)
	r.On("GetExactMatch", depToInstall).Return(&singleEntry, nil)
	allRegistries := map[string]registry.Registry{
		remoteHttpRegistry.URL: r,
	}

	var expectedListOfInstalledCommands []dependency.DependenciesIndexEntry
	for _, e := range localTestDependencyIndexEntries {
		expectedListOfInstalledCommands = append(expectedListOfInstalledCommands, dependency.DependenciesIndexEntry{
			Path: filepath.Join(validProjectInstallPath, e),
		})
	}

	mts := managerTestingSuite{
		Paths:                    context.Paths{ProjectInstallDir: validProjectInstallPath},
		DefaultRegistry:          "",
		RemoteHttpRegistryClient: remoteHttpRegistry.Client(),
		MockRegistries:           allRegistries,
		DepToUpdate:              depToUpdate,
		DepToInstall:             depToInstall,
		IsDependencyServerFaulty: false,
		InstallDir:               validProjectInstallPath,
		InstallLock:              newMockLock,

		ExpectedDepsInstalledLocally: expectedListOfInstalledCommands,
		ExpectedUpdates:              Updates{Breaking: singleEntry.Version, NonBreaking: ""},
		ExpectedInstalledDependency: dependency.Dependency{
			Name:     dependencyName,
			Registry: remoteHttpRegistry.URL,
			Version:  "2.12.1",
			Alias:    dependencyName,
		},
		CheckForUpdatesShouldFailWith: &fs.PathError{
			Op: "open", Err: fmt.Errorf(""),
		},
		CommandInstallShouldFailWith: &CantFindExactVersionMatchError{dependencyName, "2.12.1", ""},
	}

	suite.Run(t, &mts)
}

func TestInstallMajorUpdateWithNonExistentHttpRegistry(t *testing.T) {
	remoteHttpRegistry := httptest.NewServer(&testHandler{})
	defer remoteHttpRegistry.Close()

	depToUpdate := dependency.Dependency{Name: dependencyName, Registry: "http://fake.registry.io", Alias: dependencyName, Version: "2.11.2"}
	depToInstall := dependency.Dependency{Name: dependencyName, Registry: "http://fake.registry.io", Alias: dependencyName, Version: "2.12.1"}
	singleEntry := registry.Entry{
		Name:    dependencyName,
		Version: "2.12.1",
		OS:      "windows",
		Arch:    "amd64",
		URL:     fmt.Sprintf("%s/%s/%s.tar.gz", remoteHttpRegistry.URL, "registry/commands", dependencyName),
	}

	r := new(mockRegistry)
	r.On("GetHighestNonBreaking", depToUpdate).Return(&registry.Entry{}, fmt.Errorf("no non-breaking change avaialable"))
	r.On("GetHighestBreaking", depToUpdate).Return(&singleEntry, nil)
	r.On("GetExactMatch", depToInstall).Return(&singleEntry, nil)
	allRegistries := map[string]registry.Registry{
		remoteHttpRegistry.URL: r,
	}

	var expectedListOfInstalledCommands []dependency.DependenciesIndexEntry
	for _, e := range localTestDependencyIndexEntries {
		expectedListOfInstalledCommands = append(expectedListOfInstalledCommands, dependency.DependenciesIndexEntry{
			Path: filepath.Join(validProjectInstallPath, e),
		})
	}

	mts := managerTestingSuite{
		Paths:                    context.Paths{ProjectInstallDir: validProjectInstallPath},
		DefaultRegistry:          remoteHttpRegistry.URL,
		RemoteHttpRegistryClient: remoteHttpRegistry.Client(),
		MockRegistries:           allRegistries,
		DepToUpdate:              depToUpdate,
		DepToInstall:             depToInstall,
		IsDependencyServerFaulty: false,
		InstallDir:               validProjectInstallPath,
		InstallLock:              newMockLock,

		ExpectedDepsInstalledLocally: expectedListOfInstalledCommands,
		ExpectedUpdates:              Updates{Breaking: singleEntry.Version, NonBreaking: ""},
		ExpectedInstalledDependency: dependency.Dependency{
			Name:     dependencyName,
			Registry: remoteHttpRegistry.URL,
			Version:  "2.12.1",
			Alias:    dependencyName,
		},
		CheckForUpdatesShouldFailWith: &fs.PathError{Op: "open", Path: "http://fake.registry.io", Err: fmt.Errorf("")},
		CommandInstallShouldFailWith:  &CantFindExactVersionMatchError{dependencyName, "2.12.1", "http://fake.registry.io"},
	}

	suite.Run(t, &mts)
}
