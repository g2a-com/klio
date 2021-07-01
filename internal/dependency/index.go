package dependency

import "github.com/g2a-com/klio/internal/configfile"

type DependenciesIndex struct {
	Meta       configfile.Metadata      `json:"-"`
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

// SaveIndex saves a dependencies index file.
func SaveIndex(config *DependenciesIndex) error {
	return configfile.Save(config, config.Meta.Path)
}

// LoadIndex reads a dependencies index file.
func LoadIndex(filePath string) (*DependenciesIndex, error) {
	config := &DependenciesIndex{}
	if err := configfile.Load(config, &config.Meta, filePath); err != nil {
		return nil, err
	}
	return config, nil
}
