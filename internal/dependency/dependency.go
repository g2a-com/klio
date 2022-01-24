package dependency

import (
	"github.com/g2a-com/klio/internal/config"
)

// Dependency describes project's dependency - command or plugin.
type Dependency struct {
	Name     string `yaml:"name,omitempty"`
	Registry string `yaml:"registry,omitempty"`
	Version  string `yaml:"version"`
	Alias    string `yaml:"-"`
}

type DependenciesIndex struct {
	Meta       config.Metadata          `json:"-"`
	APIVersion string                   `json:"apiVersion,omitempty"`
	Kind       string                   `json:"kind,omitempty"`
	Entries    []DependenciesIndexEntry `json:"entries"`
}

type DependenciesIndexEntry struct {
	Alias    string `json:"alias"`
	Registry string `json:"registry"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Checksum string `json:"checksum"`
	Path     string `json:"path"`
}

// LoadDependenciesIndex reads a dependencies index file.
func LoadDependenciesIndex(filePath string) (*DependenciesIndex, error) {
	depConfig := &DependenciesIndex{}
	if err := config.LoadConfigFile(depConfig, &depConfig.Meta, filePath); err != nil {
		return nil, err
	}
	return depConfig, nil
}

// SaveDependenciesIndex saves a dependencies index file.
func SaveDependenciesIndex(depConfig *DependenciesIndex) error {
	return config.SaveConfigFile(depConfig, depConfig.Meta.Path)
}
