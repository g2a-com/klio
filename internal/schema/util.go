package schema

import (
	"fmt"
	"os"

	"github.com/g2a-com/klio/internal/config"
)

// LoadProjectConfig reads a project configuration file.
func LoadProjectConfig(filePath string) (*ProjectConfig, error) {
	projectConfig := &ProjectConfig{}
	if err := config.LoadConfigFile(projectConfig, &projectConfig.Meta, filePath); err != nil {
		return nil, err
	}
	return projectConfig, nil
}

// SaveProjectConfig saves a project configuration file.
func SaveProjectConfig(projectConfig *ProjectConfig) error {
	return config.SaveConfigFile(projectConfig, projectConfig.Meta.Path)
}

// LoadCommandConfig reads a project configuration file.
func LoadCommandConfig(filePath string) (*CommandConfig, error) {
	commandConfig := &CommandConfig{}
	if err := config.LoadConfigFile(commandConfig, &commandConfig.Meta, filePath); err != nil {
		return nil, err
	}
	return commandConfig, nil
}

// CreateDefaultProjectConfig creates default ProjectConfig and save it to give path if it's not already there.
func CreateDefaultProjectConfig(filePath string) (*ProjectConfig, error) {
	// create default ProjectConfig
	projectConfig := NewDefaultProjectConfig()

	// marshal newly created ProjectConfig
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

	// save ProjectConfig
	err = config.SaveConfigFile(marshaledProjectConfig, filePath)
	if err != nil {
		return projectConfig, fmt.Errorf("failed to save file in %s because of: %s", filePath, err)
	}

	return projectConfig, nil
}
