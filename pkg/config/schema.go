package config

import "gopkg.in/yaml.v3"

// CommandConfigKind defines kind property value of for g2a.yaml files
const CommandConfigKind string = "Command"

// CommandConfig describes structure of g2a.yaml files
type CommandConfig struct {
	// Meta stores metadata of the config file (such as a path).
	Meta Metadata `yaml:"-"`
	// APIVersion can be used to handle more than one config file format
	APIVersion string `yaml:"apiVersion,omitempty"`
	// Kind of the config file, currently is not very usefull since config files
	// are always placed under specified directories, but can be useful in the
	// future.
	Kind string `yaml:"kind,omitempty" validate:"eq=Command"`
	// Name of the command.
	Name string `yaml:"name,omitempty"`
	// BinPath stores *relative* path for the binary which is an entrypoint to the
	// command.
	BinPath string `yaml:"binPath,omitempty" validate:"required,file"`
	// Version of the command, semver without "v" prefix (e.g.: 1.2.3)
	Version string `yaml:"version,omitempty"`
	// Description of the command used by core "g2a" binary in order to show usage.
	Description string `yaml:"description,omitempty"`
}

// ProjectConfig describes structure of g2a.yaml files
type ProjectConfig struct {
	Meta Metadata `yaml:"-"`
	// FIXME: Find more clever way to handle this "extra" properties, we should
	// keep all properties not specified in this struct.
	APIVersion   string    `yaml:"apiVersion,omitempty"`
	Kind         string    `yaml:"kind,omitempty"`
	Name         string    `yaml:"name,omitempty"`
	Services     []string  `yaml:"services,omitempty"`
	Environments []string  `yaml:"environments,omitempty"`
	Scripts      yaml.Node `yaml:"scripts,omitempty"`

	CLI CLIConfig `yaml:"cli,omitempty"`
}

// Metadata contains additional info, such as path of config file
type Metadata struct {
	Path string
}

// CLIConfig contains configuration for CLI itself
type CLIConfig struct {
	Commands map[string]string `yaml:"commands,omitempty"`
}
