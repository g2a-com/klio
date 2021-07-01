package configfile

// Metadata contains additional info, such as path of config file
type Metadata struct {
	Path   string
	Exists bool
}

type Generic struct {
	// Meta stores metadata of the config file (such as a path).
	Meta Metadata `yaml:"-"`
	// APIVersion can be used to handle more than one config file format
	APIVersion string `yaml:"apiVersion"`
	// Kind of the config file
	Kind string `yaml:"kind"`
}

// PluginConfig describes structure of klio.yaml files
type PluginConfig struct {
	// Meta stores metadata of the config file (such as a path).
	Meta Metadata `yaml:"-"`
	// APIVersion can be used to handle more than one config file format
	APIVersion string `yaml:"apiVersion,omitempty"`
	// Kind of the config file
	Kind string `yaml:"kind,omitempty" validate:"eq=Command"`
	// Name of the command.
	BinPath string `yaml:"binPath,omitempty" validate:"required,file"`
	// Description of the command used by core "klio" binary in order to show usage.
	Description string `yaml:"description,omitempty"`
}
