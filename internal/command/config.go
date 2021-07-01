package command

import "github.com/g2a-com/klio/internal/configfile"

// Kind defines kind property value of for klio.yaml files
const Kind string = "Command"

// Config describes structure of klio.yaml files
type Config struct {
	// Meta stores metadata of the config file (such as a path).
	Meta configfile.Metadata `yaml:"-"`
	// APIVersion can be used to handle more than one config file format
	APIVersion string `yaml:"apiVersion,omitempty"`
	// Kind of the config file
	Kind string `yaml:"kind,omitempty" validate:"eq=Command"`
	// Name of the command.
	BinPath string `yaml:"binPath,omitempty" validate:"required,file"`
	// Description of the command used by core "klio" binary in order to show usage.
	Description string `yaml:"description,omitempty"`
}

// LoadConfig reads a project configuration file.
func LoadConfig(filePath string) (*Config, error) {
	config := &Config{}
	if err := configfile.Load(config, &config.Meta, filePath); err != nil {
		return nil, err
	}
	return config, nil
}
