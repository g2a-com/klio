package cmd

import "github.com/g2a-com/klio/internal/config"

// Config describes structure of klio.yaml files.
type Config struct {
	// Meta stores metadata of the config file (such as a path).
	Meta config.Metadata `yaml:"-"`
	// APIVersion can be used to handle more than one config file format
	APIVersion string `yaml:"apiVersion,omitempty"`
	// Kind of the config file
	Kind string `yaml:"kind,omitempty" validate:"eq=Command"`
	// Name of the command.
	BinPath string `yaml:"binPath,omitempty" validate:"required,file"`
	// Description of the command used by core "klio" binary in order to show usage.
	Description string `yaml:"description,omitempty"`
	// Version of currently installed command
	Version string `yaml:"version,omitempty"`
}

// LoadConfig reads a command configuration file.
func LoadConfig(filePath string) (*Config, error) {
	commandConfig := &Config{}
	if err := config.LoadConfigFile(commandConfig, &commandConfig.Meta, filePath); err != nil {
		return nil, err
	}
	return commandConfig, nil
}
