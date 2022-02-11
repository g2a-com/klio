package manager

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/dependency/registry"
	"github.com/g2a-com/klio/internal/lock"
	"github.com/spf13/afero"
	"io"
	"net/http"
	"path/filepath"
)

const (
	validProjectInstallPath = "valid/project/install/path"
	validGlobalInstallPath  = "valid/global/install/path"

	dependencyName = "dosomething"
)

//=========================== DEP INDEX ===========================

func fetchInvalidIndex(_ string) (*dependency.DependenciesIndex, error) {
	return nil, fmt.Errorf("an error")
}

var simpleValidIndex = dependency.DependenciesIndex{
	Entries: []dependency.DependenciesIndexEntry{
		{
			Path: "aaa/bbb",
		},
		{
			Path: "ccc/ddd",
		},
	},
}

func fetchSimpleValidIndex(file string) (*dependency.DependenciesIndex, error) {
	if file != filepath.Join(validProjectInstallPath, "dependencies.json") {
		return nil, fmt.Errorf("an error")
	}
	return &simpleValidIndex, nil
}

var properlyProcessedValidSimpleIndex = []dependency.DependenciesIndexEntry{
	{
		Path: filepath.Join(validProjectInstallPath, "aaa/bbb"),
	},
	{
		Path: filepath.Join(validProjectInstallPath, "ccc/ddd"),
	},
}

//=========================== REGISTRY MOCKS ===========================

const (
	noHighestVersionRegistryName = "noHighestVersion"
	noExactVersionRegistryName   = "noExactVersion"
	regularRegistryName          = "fairlyRegular"
	borderVersion                = "3.0.2"
)

var allRegistries = map[string]registry.Registry{
	noHighestVersionRegistryName: &mockRegistry{getHighestError: fmt.Errorf("highest error")},
	noExactVersionRegistryName:   &mockRegistry{getExactError: fmt.Errorf("exact error")},
	regularRegistryName:          &mockRegistry{},
}

type mockRegistry struct {
	registryURL     string
	getHighestError error
	getExactError   error
	wrongChecksum   bool
}

func (r *mockRegistry) Update() error {
	return nil
}

//bump major If new version below border
func (r *mockRegistry) GetHighestBreaking(dep dependency.Dependency) (*registry.Entry, error) {
	if r.getHighestError != nil {
		return nil, r.getHighestError
	}
	baseVersion, _ := semver.NewVersion(dep.Version)
	newVersion := baseVersion.IncMajor()
	borderVersion, _ := semver.NewVersion(borderVersion)

	if newVersion.LessThan(borderVersion) {
		return &registry.Entry{
			Name:        dep.Name,
			Version:     newVersion.String(),
			OS:          "windows",
			Arch:        "amd64",
			Annotations: nil,
		}, nil
	}
	return nil, nil
}

//add 1 to patch If new version below border
func (r *mockRegistry) GetHighestNonBreaking(dep dependency.Dependency) (*registry.Entry, error) {
	if r.getHighestError != nil {
		return nil, r.getHighestError
	}
	baseVersion, _ := semver.NewVersion(dep.Version)
	newVersion := baseVersion.IncPatch()
	borderVersion, _ := semver.NewVersion(borderVersion)

	if baseVersion.LessThan(borderVersion) {
		return &registry.Entry{
			Name:        dep.Name,
			Version:     newVersion.String(),
			OS:          "windows",
			Arch:        "amd64",
			Annotations: nil,
		}, nil
	}
	return nil, nil
}

//=========================== REGISTRY SERVER STUB ===========================

const (
	validArtifactoryPath      = "valid/artifactory/path"
	firstRegistryStubVersion  = "1.1.0"
	secondRegistryStubVersion = "2.1.0"
)

var tarPathToDependencyOnArtifactory = fmt.Sprintf("/give/me/my/tar/%s.tar.gz", dependencyName)

type testHandler struct{}

func (th *testHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.URL.Path != tarPathToDependencyOnArtifactory {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	//hard dependency in resources
	file, err := afero.NewOsFs().Open(fmt.Sprintf("%s.tar.gz", dependencyName))
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	all, err := io.ReadAll(file)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = rw.Write(all)
}

//this part of mock registry is dedicated to the stub
func (r *mockRegistry) GetExactMatch(dep dependency.Dependency) (*registry.Entry, error) {
	if r.getExactError != nil || dep.Name != dependencyName {
		return nil, r.getExactError
	}

	rrr := registry.Entry{
		Name:        dependencyName,
		Version:     dep.Version,
		OS:          "windows",
		Arch:        "amd64",
		Annotations: nil,
		URL:         fmt.Sprintf("%s%s", r.registryURL, tarPathToDependencyOnArtifactory),
		Checksum:    "sha256-5681f156ae3b120c63451a2e648e17914429315c5c3d7e78e892093f530ca853",
	}

	if r.wrongChecksum {
		rrr.Checksum = "sha256-somereallyspoiledchecksum"
	}

	return &rrr, nil
}

var depIndexForRegistryStub = dependency.DependenciesIndex{
	Entries: []dependency.DependenciesIndexEntry{
		{
			Name:    dependencyName,
			Path:    "dependency/second",
			Version: secondRegistryStubVersion,
		},
		{
			Name:    dependencyName,
			Path:    "dependency/first",
			Version: firstRegistryStubVersion,
		},
	},
}

func fetchDepIndexForRegistryStub(file string) (*dependency.DependenciesIndex, error) {
	if file != filepath.Join(validProjectInstallPath, "dependencies.json") {
		return nil, fmt.Errorf("wrong path according to tests")
	}
	return &depIndexForRegistryStub, nil
}

func mockSave(_ *dependency.DependenciesIndex) error         { return nil }
func mockMessedUpSave(_ *dependency.DependenciesIndex) error { return fmt.Errorf("some save error") }

//=========================== LOCK ===========================

type mockLock struct{}

func newMockLock(_ string) (lock.Lock, error) { return &mockLock{}, nil }

func (ml *mockLock) Acquire() error { return nil }

func (ml *mockLock) Release() error { return nil }

//=========================== LOCAL FS ===========================

var fsPaths = []string{
	validArtifactoryPath,
	"some/path/nobody/cares/about",
}

func getMockFs() afero.Fs {
	fs := afero.NewMemMapFs()

	for _, p := range fsPaths {
		_ = fs.MkdirAll(p, 0x777)
	}

	return fs
}
