package project

import (
	"fmt"
	"github.com/g2a-com/klio/internal/dependency"
	"gopkg.in/yaml.v3"
	"os"

	"github.com/g2a-com/klio/internal/configfile"
)

type Project interface {
	GetDependencies() []dependency.Dependency
	GetDefaultRegistry() string

	SetDependencies([]dependency.Dependency)
	SetDefaultRegistry(string)

	SaveConfig() error
}

// config describes structure of klio.yaml files
type config struct {
	Meta            configfile.Metadata
	DefaultRegistry string
	Dependencies    []dependency.Dependency
	yaml            *yaml.Node
}

// LoadConfig reads a project configuration file.
func LoadConfig(filePath string) (Project, error) {
	config := &config{}
	if err := configfile.Load(config, &config.Meta, filePath); err != nil {
		return nil, err
	}
	return config, nil
}

// CreateDefaultConfig creates default ProjectConfig and save it to give path if it's not already there
func CreateDefaultConfig(filePath string) (Project, error) {
	// create default ProjectConfig
	projectConfig := newDefaultProjectConfig()

	// marshal newly created ProjectConfig
	marshaledProjectConfig, err := projectConfig.marshalYAML()
	if err != nil {
		return projectConfig, fmt.Errorf("failed to marshal: %s", err)
	}

	// check if file already exists, if so return error
	_, err = os.Stat(filePath)
	isFileNotExist := os.IsNotExist(err)
	if !isFileNotExist {
		return projectConfig, fmt.Errorf("failed to create klio.yaml file, it already exists at %s", filePath)
	}

	// save ProjectConfig
	err = configfile.Save(marshaledProjectConfig, filePath)
	if err != nil {
		return projectConfig, fmt.Errorf("failed to save file in %s because of: %s", filePath, err)
	}

	return projectConfig, nil
}

// SaveConfig saves a project configuration file.
func (p *config) SaveConfig() error {
	return configfile.Save(p, p.Meta.Path)
}

func (p *config) GetDefaultRegistry() string {
	return p.DefaultRegistry
}

func (p *config) GetDependencies() []dependency.Dependency {
	return p.Dependencies
}

func (p *config) SetDefaultRegistry(registry string) {
	p.DefaultRegistry = registry
}

func (p *config) SetDependencies(deps []dependency.Dependency) {
	p.Dependencies = deps
}
