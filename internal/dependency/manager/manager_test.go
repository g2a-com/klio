package manager

import (
	"github.com/g2a-com/klio/internal/context"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/spf13/afero"
	"net/http/httptest"
	"reflect"
	"testing"
)

//==========================================================================================
//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
//==========================================================================================
// Take a look at DEP INDEX in utils_test.go
//==========================================================================================

func TestGetInstalledCommands(t *testing.T) {
	tests := []struct {
		// Name of the test
		name string
		// Injected mock function for returning local dependencies
		fetchIndexFunction func(filePath string) (*dependency.DependenciesIndex, error)
		// Configuration paths for klio
		paths context.Paths
		// Desired index list of dependencies installed
		want []dependency.DependenciesIndexEntry
	}{
		{
			name:               "errorFromIndex",
			fetchIndexFunction: fetchInvalidIndex,
			paths: context.Paths{
				ProjectInstallDir: validProjectInstallPath,
				GlobalInstallDir:  validGlobalInstallPath,
			},
			want: nil,
		},
		{
			name:               "emptyDirConfig",
			fetchIndexFunction: fetchInvalidIndex,
			paths: context.Paths{
				ProjectInstallDir: "",
				GlobalInstallDir:  "",
			},
			want: nil,
		},
		{
			name:               "simpleValidIndex",
			fetchIndexFunction: fetchSimpleValidIndex,
			paths: context.Paths{
				ProjectInstallDir: validProjectInstallPath,
				GlobalInstallDir:  validGlobalInstallPath,
			},
			want: properlyProcessedValidSimpleIndex,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := &Manager{
				fetchDependencyIndex: tt.fetchIndexFunction,
			}
			if got := mgr.GetInstalledCommands(tt.paths); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetInstalledCommands() = %v, want %v", got, tt.want)
			}
		})
	}
}

//==========================================================================================
//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
//==========================================================================================
// Take a look at REGISTRY MOCKS in utils_test.go
//==========================================================================================

func TestGetUpdateFor(t *testing.T) {
	tests := []struct {
		// Name of the test
		name string
		// Dependency to get update for
		dep dependency.Dependency
		// Desired list of versions available
		want Updates
		// Flag marking that the test should end up with error
		wantErr bool
	}{
		{
			name: "nonExistentRegistry",
			dep: dependency.Dependency{
				Registry: "nonExistent",
			},
			want:    Updates{},
			wantErr: true,
		},
		{
			name: "fetchHighestMalfunction", //non intrusive error
			dep: dependency.Dependency{
				Name:     dependencyName,
				Registry: noHighestVersionRegistryName,
				Version:  "2.1.0",
			},
			want:    Updates{},
			wantErr: false,
		},
		{
			name: "fetch2.1.0",
			dep: dependency.Dependency{
				Name:     dependencyName,
				Registry: regularRegistryName,
				Version:  "2.1.0",
			},
			want:    Updates{"2.1.1", "3.0.0"},
			wantErr: false,
		},
		{
			name: "fetch2.7.0",
			dep: dependency.Dependency{
				Name:     dependencyName,
				Registry: regularRegistryName,
				Version:  "2.7.0",
			},
			want:    Updates{"2.7.1", "3.0.0"},
			wantErr: false,
		},
		{
			name: "fetch3.0.0",
			dep: dependency.Dependency{
				Name:     dependencyName,
				Registry: regularRegistryName,
				Version:  "3.0.0",
			},
			want:    Updates{"3.0.1", ""},
			wantErr: false,
		},
		{
			name: "fetch3.1.0",
			dep: dependency.Dependency{
				Name:     dependencyName,
				Registry: regularRegistryName,
				Version:  "3.1.0",
			},
			want:    Updates{"", ""},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := &Manager{
				DefaultRegistry: "",
				registries:      allRegistries,
				os:              afero.NewMemMapFs(),
			}
			got, err := mgr.GetUpdateFor(tt.dep)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUpdateFor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUpdateFor() got = %v, want %v", got, tt.want)
			}
		})
	}
}

//==========================================================================================
//::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
//==========================================================================================
// Take a look at REGISTRY SERVER STUB in utils_test.go
//==========================================================================================

func TestManagerInstallDependency(t *testing.T) {
	handler := &testHandler{}
	validServer := httptest.NewServer(handler)
	defer validServer.Close()
	invalidChecksumServer := httptest.NewServer(handler)
	defer invalidChecksumServer.Close()

	//extending registries with a mock http registry
	allRegistries[validServer.URL] = &mockRegistry{registryURL: validServer.URL}
	allRegistries[invalidChecksumServer.URL] = &mockRegistry{registryURL: invalidChecksumServer.URL, wrongChecksum: true}

	tests := []struct {
		// Name of the test
		name string
		// Default registry to query for dependency
		defaultRegistry string
		// Dependency to install
		dep dependency.Dependency
		// Directory to install the dependency in
		installDir string
		//function that saves dependency
		saveFunction func(depConfig *dependency.DependenciesIndex) error
		// Desired dependency structure that was mock installed
		want dependency.Dependency
		// Flag marking that the test should end up with error
		wantErr bool
	}{
		{
			name:            "validExistingDep",
			defaultRegistry: validServer.URL,
			dep: dependency.Dependency{
				Name:     dependencyName,
				Registry: validServer.URL,
				Version:  secondRegistryStubVersion,
			},
			installDir:   validProjectInstallPath,
			saveFunction: mockSave,
			want: dependency.Dependency{
				Name:     dependencyName,
				Registry: validServer.URL,
				Version:  secondRegistryStubVersion,
				Alias:    dependencyName,
			},
			wantErr: false,
		},
		{
			name:            "invalidNonexistingDep",
			defaultRegistry: validServer.URL,
			dep: dependency.Dependency{
				Name:     "someNonexistentDep",
				Registry: validServer.URL,
				Version:  "2.2.0",
			},
			installDir:   validProjectInstallPath,
			saveFunction: mockSave,
			wantErr:      true,
		},
		{
			name:            "invalidChecksum",
			defaultRegistry: invalidChecksumServer.URL,
			dep: dependency.Dependency{
				Name:     dependencyName,
				Registry: invalidChecksumServer.URL,
				Version:  "2.2.0",
			},
			installDir:   validProjectInstallPath,
			saveFunction: mockSave,
			wantErr:      true,
		},
		{
			name:            "deadArtifactory",
			defaultRegistry: regularRegistryName,
			dep: dependency.Dependency{
				Name:     dependencyName,
				Registry: regularRegistryName,
				Version:  "2.2.0",
			},
			installDir:   validProjectInstallPath,
			saveFunction: mockSave,
			wantErr:      true,
		},
		{
			name:            "messedUpSave",
			defaultRegistry: validServer.URL,
			dep: dependency.Dependency{
				Name:     dependencyName,
				Registry: validServer.URL,
				Version:  "2.2.0",
			},
			installDir:   validProjectInstallPath,
			saveFunction: mockMessedUpSave,
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := &Manager{
				DefaultRegistry:      tt.defaultRegistry,
				registries:           allRegistries,
				httpDownloadClient:   validServer.Client(),
				os:                   getMockFs(),
				createLock:           newMockLock,
				fetchDependencyIndex: fetchDepIndexForRegistryStub,
				saveIndex:            tt.saveFunction,
			}
			err := mgr.InstallDependency(&tt.dep, tt.installDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("InstallDependency() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				t.Logf("error: %s", err)
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.dep, tt.want) {
				t.Errorf("InstallDependency() got = %v, want %v", tt.dep, tt.want)
			}
		})
	}
}
