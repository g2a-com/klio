package schema

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v3"

	"github.com/g2a-com/klio/internal/log"
)

// LoadConfigFile reads, parses and validates specified configuration file
func LoadConfigFile(dataStruct interface{}, meta *Metadata, configFilePath string) error {
	log.Spamf(`Loading file "%s"...`, configFilePath)

	absPath, err := filepath.Abs(configFilePath)
	if err != nil {
		return err
	}

	buf, err := ioutil.ReadFile(absPath)
	if err == nil {
		switch ext := path.Ext(absPath); ext {
		case ".yaml", ".yml":
			if err := yaml.Unmarshal(buf, dataStruct); err != nil {
				return err
			}
		case ".json":
			if err := json.Unmarshal(buf, dataStruct); err != nil {
				return err
			}
		default:
			return fmt.Errorf(`unsupported extension "%s" of file: %s`, ext, absPath)
		}

		meta.Exists = true
	} else if !os.IsNotExist(err) {
		return err
	}

	meta.Path = absPath

	if err := validate(dataStruct, filepath.Dir(configFilePath)); err != nil {
		return fmt.Errorf(`file "%s" doesn't pass validation: %s`, configFilePath, err)
	}

	return nil
}

// SaveConfigFile validates, serializes and saves configuration to a file
func SaveConfigFile(data interface{}, configFilePath string) error {
	log.Spamf(`Saving file "%s"...`, configFilePath)

	if err := validate(data, filepath.Dir(configFilePath)); err != nil {
		return fmt.Errorf(`failed to generate valid config for %s: %s`, configFilePath, err)
	}

	file, err := os.OpenFile(configFilePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	switch ext := path.Ext(configFilePath); ext {
	case ".yaml", ".yml":
		encoder := yaml.NewEncoder(file)
		encoder.SetIndent(2)
		err = encoder.Encode(data)
	case ".json":
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		err = encoder.Encode(data)
	default:
		return fmt.Errorf(`unsupported extension "%s" of file: %s`, ext, configFilePath)
	}

	if err != nil {
		return err
	}

	return nil
}

// LoadProjectConfig reads a project configuration file.
func LoadProjectConfig(filePath string) (*ProjectConfig, error) {
	config := &ProjectConfig{}
	if err := LoadConfigFile(config, &config.Meta, filePath); err != nil {
		return nil, err
	}
	return config, nil
}

// SaveProjectConfig saves a project configuration file.
func SaveProjectConfig(config *ProjectConfig) error {
	return SaveConfigFile(config, config.Meta.Path)
}

// LoadCommandConfig reads a project configuration file.
func LoadCommandConfig(filePath string) (*CommandConfig, error) {
	config := &CommandConfig{}
	if err := LoadConfigFile(config, &config.Meta, filePath); err != nil {
		return nil, err
	}
	return config, nil
}

// LoadDependenciesIndex reads a dependenceis index file.
func LoadDependenciesIndex(filePath string) (*DependenciesIndex, error) {
	config := &DependenciesIndex{}
	if err := LoadConfigFile(config, &config.Meta, filePath); err != nil {
		return nil, err
	}
	return config, nil
}

// SaveDependenciesIndex saves a dependenceis index file.
func SaveDependenciesIndex(config *DependenciesIndex) error {
	return SaveConfigFile(config, config.Meta.Path)
}

// CreateDefaultProjectConfig creates default ProjectConfig and save it to give path if it's not already there
func CreateDefaultProjectConfig(filePath string) (*ProjectConfig, error) {
	// create default ProjectConfig
	projectConfig, err := NewDefaultProjectConfig()
	if err != nil {
		log.Fatalf("Failed to unmarshal default yaml: %s", err)
	}

	// marshal newly created ProjectConfig
	marshaledProjectConfig, err := projectConfig.MarshalYAML()
	if err != nil {
		log.Fatalf("Failed to marshal: %s", err)
	}

	// check if file already exists, if so return error
	_, err = os.Stat(filePath)
	isFileNotExist := os.IsNotExist(err)
	if !isFileNotExist {
		return projectConfig, fmt.Errorf("Failed to create klio.yaml file, it already exists at %s", filePath)
	}

	// save ProjectConfig
	err = SaveConfigFile(marshaledProjectConfig, filePath)
	if err != nil {
		log.Fatalf("Failed to save file in %s because of: %s", filePath, err)
	}

	return projectConfig, nil
}

func validate(data interface{}, dir string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	validate := validator.New()

	os.Chdir(dir)
	err = validate.Struct(data)
	os.Chdir(cwd)

	if err != nil {
		fieldError := err.(validator.ValidationErrors)[0]
		tag := fieldError.Tag()
		if fieldError.Param() != "" {
			tag += "=" + fieldError.Param()
		}
		return fmt.Errorf(`field validation for %s failed on the %s`, fieldError.Namespace(), tag)
	}

	return nil
}
