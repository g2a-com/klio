package project

import (
	"errors"

	"github.com/g2a-com/klio/internal/config"
	"github.com/g2a-com/klio/internal/dependency"
	"gopkg.in/yaml.v3"
)

// Config describes structure of klio.yaml files.
type Config struct {
	Meta            config.Metadata
	DefaultRegistry string
	Dependencies    []dependency.Dependency
	yaml            *yaml.Node
}

func NewDefaultConfig() *Config {
	projectConfig := Config{}

	_ = projectConfig.UnmarshalYAML(minimalConfig())

	return &projectConfig
}

func (p Config) MarshalYAML() (interface{}, error) {
	var defaultRegistryValueNode *yaml.Node
	var dependenciesValueNode *yaml.Node

	if p.yaml == nil {
		return minimalConfig(), nil
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
	_ = defaultRegistryValueNode.Encode(p.DefaultRegistry)

	// Encode dependencies
	dependencies := map[string]dependency.Dependency{}
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
	_ = dependenciesValueNode.Encode(dependencies)

	// Return result
	return p.yaml, nil
}

func minimalConfig() *yaml.Node {
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

func (p *Config) UnmarshalYAML(node *yaml.Node) error {
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
			_ = v.Decode(&p.DefaultRegistry)
		case "dependencies":
			aux := map[string]dependency.Dependency{}
			_ = v.Decode(&aux)
			for alias, dep := range aux {
				dep.Alias = alias
				p.Dependencies = append(p.Dependencies, dep)
			}
		}
	}

	// Preserve original YAML data
	p.yaml = node

	// Normalize dependencies
	for i := range p.Dependencies {
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
