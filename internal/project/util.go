package project

import (
	"fmt"
	"os"

	"github.com/g2a-com/klio/internal/config"
)

// LoadProjectConfig reads a project configuration file.
func LoadProjectConfig(filePath string) (*Config, error) {
	projectConfig := &Config{}
	if err := config.LoadConfigFile(projectConfig, &projectConfig.Meta, filePath); err != nil {
		return nil, err
	}
	return projectConfig, nil
}

// SaveProjectConfig saves a project configuration file.
func SaveProjectConfig(projectConfig *Config) error {
	return config.SaveConfigFile(projectConfig, projectConfig.Meta.Path)
}

// CreateDefaultProjectConfig creates default Config and save it to give path if it's not already there.
func CreateDefaultProjectConfig(filePath string) (*Config, error) {
	// create default Config
	projectConfig := NewDefaultConfig()

	// marshal newly created Config
	marshaledProjectConfig, err := projectConfig.MarshalYAML()
	if err != nil {
		return projectConfig, fmt.Errorf("failed to marshal: %s", err)
	}

	// check if file already exists, if so return error
	_, err = os.Stat(filePath)
	isFileNotExist := os.IsNotExist(err)
	if !isFileNotExist {
		return projectConfig, fmt.Errorf("failed to create klio.yaml file, it already exists at %s", filePath)
	}

	// save Config
	err = config.SaveConfigFile(marshaledProjectConfig, filePath)
	if err != nil {
		return projectConfig, fmt.Errorf("failed to save file in %s because of: %s", filePath, err)
	}

	return projectConfig, nil
}
