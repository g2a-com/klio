package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/g2a-com/klio/internal/log"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v3"
)

// Metadata contains context and additional info for configuration files.
type Metadata struct {
	Path   string
	Exists bool
}

type GenericConfigFile struct {
	// Meta stores metadata of the config file (such as a path).
	Meta Metadata `yaml:"-"`
	// APIVersion can be used to handle more than one config file format
	APIVersion string `yaml:"apiVersion"`
	// Kind of the config file
	Kind string `yaml:"kind"`
}

// LoadConfigFile reads, parses and validates specified configuration file.
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

// SaveConfigFile validates, serializes and saves configuration to a file.
func SaveConfigFile(data interface{}, configFilePath string) error {
	log.Spamf(`Saving file "%s"...`, configFilePath)

	if err := validate(data, filepath.Dir(configFilePath)); err != nil {
		return fmt.Errorf(`failed to generate valid config for %s: %s`, configFilePath, err)
	}

	file, err := os.OpenFile(configFilePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
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

func validate(data interface{}, dir string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	validate := validator.New()

	_ = os.Chdir(dir)
	err = validate.Struct(data)
	_ = os.Chdir(cwd)

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
