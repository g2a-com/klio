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

// SetDefaults puts default values for registry for alias and registry (if missing).
func (dep *Dependency) SetDefaults(defaultRegistry string) {
	// Fill missing values with defaults
	if dep.Alias == "" {
		dep.Alias = dep.Name
	}
	if dep.Registry == "" {
		dep.Registry = defaultRegistry
	}
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

type IndexHandler interface {
	LoadDependencyIndex(filePath string) error
	SaveDependencyIndex() error
	GetEntries() []DependenciesIndexEntry
	SetEntries(entries []DependenciesIndexEntry)
}

type LocalIndexHandler struct {
	dependencyIndex *DependenciesIndex
}

// LoadDependencyIndex reads a dependencies index file.
func (di *LocalIndexHandler) LoadDependencyIndex(filePath string) error {
	depConfig := &DependenciesIndex{}
	if err := config.LoadConfigFile(depConfig, &depConfig.Meta, filePath); err != nil {
		return err
	}
	di.dependencyIndex = depConfig
	return nil
}

// SaveDependencyIndex saves a dependencies index file.
func (di *LocalIndexHandler) SaveDependencyIndex() error {
	return config.SaveConfigFile(di.dependencyIndex, di.dependencyIndex.Meta.Path)
}

func (di *LocalIndexHandler) GetEntries() []DependenciesIndexEntry {
	if di.dependencyIndex != nil {
		return di.dependencyIndex.Entries
	}
	return []DependenciesIndexEntry{}
}

func (di *LocalIndexHandler) SetEntries(entries []DependenciesIndexEntry) {
	di.dependencyIndex.Entries = entries
}
