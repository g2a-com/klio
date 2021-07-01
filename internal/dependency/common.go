package dependency

// Dependency describes project's dependency - command or plugin
type Dependency struct {
	Name     string `yaml:"name,omitempty"`
	Registry string `yaml:"registry,omitempty"`
	Version  string `yaml:"version"`
	Checksum string `yaml:"checksum"`
	Alias    string `yaml:"-"`
}
