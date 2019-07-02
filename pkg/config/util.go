package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v3"
)

// LoadConfigFile reads, parses and validates specified configuration file
func LoadConfigFile(dataStruct interface{}, configFilePath string) error {
	buf, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	switch ext := path.Ext(configFilePath); ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(buf, dataStruct); err != nil {
			return err
		}
	case ".json":
		if err := json.Unmarshal(buf, dataStruct); err != nil {
			return err
		}
	default:
		return fmt.Errorf(`unsupported extension "%s" of file: %s`, ext, configFilePath)
	}

	if err := validate(dataStruct, filepath.Dir(configFilePath)); err != nil {
		return fmt.Errorf(`invalid config "%s": %s`, configFilePath, err)
	}

	return nil
}

// SaveConfigFile validates, serializes and saves configuration to a file
func SaveConfigFile(data interface{}, configFilePath string) error {
	if err := validate(data, filepath.Dir(configFilePath)); err != nil {
		return fmt.Errorf(`failed to generate valid config for %s: %s`, configFilePath, err)
	}

	file, err := os.OpenFile(configFilePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)

	err = encoder.Encode(data)
	if err != nil {
		return err
	}

	return nil
}

// LoadProjectConfig reads a project configuration file.
func LoadProjectConfig(filePath string) (*ProjectConfig, error) {
	config := &ProjectConfig{}
	err := LoadConfigFile(config, filePath)
	if err != nil {
		return nil, err
	}
	config.Meta.Path, err = filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// LoadCommandConfig reads a project configuration file.
func LoadCommandConfig(filePath string) (*CommandConfig, error) {
	config := &CommandConfig{}
	err := LoadConfigFile(config, filePath)
	if err != nil {
		return nil, err
	}
	config.Meta.Path, err = filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// SaveProjectConfig saves a project configuration file.
func SaveProjectConfig(config *ProjectConfig) error {
	err := SaveConfigFile(config, config.Meta.Path)
	if err != nil {
		return err
	}
	return nil
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
