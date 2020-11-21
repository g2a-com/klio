package schema

import (
	"errors"

	"gopkg.in/yaml.v3"
)

// CommandConfigKind defines kind property value of for klio.yaml files
const CommandConfigKind string = "Command"

type GenericConfigFile struct {
	// Meta stores metadata of the config file (such as a path).
	Meta Metadata `yaml:"-"`
	// APIVersion can be used to handle more than one config file format
	APIVersion string `yaml:"apiVersion"`
	// Kind of the config file
	Kind string `yaml:"kind"`
}

// CommandConfig describes structure of klio.yaml files
type CommandConfig struct {
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

// ProjectConfig describes structure of klio.yaml files
type ProjectConfig struct {
	Meta            Metadata
	DefaultRegistry string
	Dependencies    []Dependency
	yaml            *yaml.Node
}

func NewDefaultProjectConfig() (*ProjectConfig, error) {
	projectConfig := ProjectConfig{}

	err := projectConfig.UnmarshalYAML(minimalKlioFile())
	if err != nil {
		return &projectConfig, err
	}

	return &projectConfig, nil
}

func (p ProjectConfig) MarshalYAML() (interface{}, error) {
	var defaultRegistryValueNode *yaml.Node
	var dependenciesValueNode *yaml.Node

	if p.yaml == nil {
		return minimalKlioFile(), nil
	}

	// Find nodes to encode
	for i := 0; i < len(p.yaml.Content)/2; i++ {
		k := p.yaml.Content[i*2]
		v := p.yaml.Content[i*2+1]

		if k.Tag != "!!str" {
			continue
		}

		switch k.Value {
		case "defaultRegistry":
			defaultRegistryValueNode = v
		case "dependencies":
			dependenciesValueNode = v
		}
	}

	// If there is no defaultRegistry node, create it
	if defaultRegistryValueNode == nil {
		defaultRegistryValueNode = &yaml.Node{}
		p.yaml.Content = append(p.yaml.Content, &yaml.Node{Value: "defaultRegistry", Tag: "!!str", Kind: yaml.ScalarNode}, defaultRegistryValueNode)
	}

	// If there is no dependencies node, create it
	if dependenciesValueNode == nil {
		dependenciesValueNode = &yaml.Node{}
		p.yaml.Content = append(p.yaml.Content, &yaml.Node{Value: "dependencies", Tag: "!!str", Kind: yaml.ScalarNode}, dependenciesValueNode)
	}

	// Encode defaultRegistry
	defaultRegistryValueNode.Encode(p.DefaultRegistry)

	// Encode dependencies
	dependencies := map[string]Dependency{}
	for _, d := range p.Dependencies {
		key := d.Alias
		d.Alias = ""
		if d.Name == key {
			d.Name = ""
		}
		if d.Registry == p.DefaultRegistry {
			d.Registry = ""
		}
		dependencies[key] = d
	}
	dependenciesValueNode.Encode(dependencies)

	// Return result
	return p.yaml, nil
}

func minimalKlioFile() *yaml.Node {
	return &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
		Content: []*yaml.Node{
			{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: "dependencies",
			},
			{
				Kind: yaml.MappingNode,
				Tag:  "!!map",
			},
		},
	}
}

func (p *ProjectConfig) UnmarshalYAML(node *yaml.Node) error {
	if node.Tag != "!!map" {
		return errors.New("invalid format")
	}

	// Unmarshal data
	for i := 0; i < len(node.Content)/2; i++ {
		k := node.Content[i*2]
		v := node.Content[i*2+1]

		if k.Tag != "!!str" {
			continue
		}

		switch k.Value {
		case "defaultRegistry":
			v.Decode(&p.DefaultRegistry)
		case "dependencies":
			aux := map[string]Dependency{}
			v.Decode(&aux)
			for alias, dep := range aux {
				dep.Alias = alias
				p.Dependencies = append(p.Dependencies, dep)
			}
		}
	}

	// Preserve original YAML data
	p.yaml = node

	// Normalize dependencies
	for i, _ := range p.Dependencies {
		d := &p.Dependencies[i]
		if d.Registry == "" {
			d.Registry = p.DefaultRegistry
		}
		if d.Name == "" {
			d.Name = d.Alias
		}
	}

	return nil
}

// Metadata contains additional info, such as path of config file
type Metadata struct {
	Path   string
	Exists bool
}

// Dependency describes project's dependency - command or plugin
type Dependency struct {
	Name     string `yaml:"name,omitempty"`
	Registry string `yaml:"registry,omitempty"`
	Version  string `yaml:"version"`
	Checksum string `yaml:"checksum"`
	Alias    string `yaml:"-"`
}

type Registry struct {
	Meta        Metadata          `yaml:"-"`
	APIVersion  string            `yaml:"apiVersion,omitempty"`
	Kind        string            `yaml:"kind,omitempty"`
	Annotations map[string]string `yaml:"annotations"`
	Entries     []RegistryEntry   `yaml:"entries"`
}

type RegistryEntry struct {
	Name        string            `yaml:"name"`
	Version     string            `yaml:"version"`
	OS          string            `json:"os"`
	Arch        string            `json:"arch"`
	Annotations map[string]string `yaml:"annotations"`
	URL         string            `yaml:"url"`
	Checksum    string            `yaml:"checksum"`
}

type RegistryEntryVersion struct {
	Number string `yaml:"number" json:"number"`
	OS     string `yaml:"os" json:"os"`
	Arch   string `yaml:"arch" json:"arch"`
}

type DependenciesIndex struct {
	Meta       Metadata                 `json:"-"`
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
