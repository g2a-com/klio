package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/yaml.v2"
)

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

	validate := validator.New()
	err = validate.Struct(dataStruct)
	if err != nil {
		fieldError := err.(validator.ValidationErrors)[0]
		tag := fieldError.Tag()
		if fieldError.Param() != "" {
			tag += "=" + fieldError.Param()
		}
		return fmt.Errorf(`invalid config "%s": field validation for %s failed on the %s`, configFilePath, fieldError.Namespace(), tag)
	}

	return nil
}
