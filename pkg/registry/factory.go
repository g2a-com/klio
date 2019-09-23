package registry

import (
	"github.com/pkg/errors"

	"github.com/g2a-com/klio/pkg/config"
)

func NewRegistriesMap(config config.CmdRegistriesConfig) (map[string]*Registry, error) {
	rm := make(map[string]*Registry)
	{
		defaultRegistry, err := New(DefaultRegistry)
		if err != nil {
			return rm, errors.Wrap(err, "failed to create registry object")
		}
		rm["default"] = defaultRegistry
	}
	if config.Artifactory != nil {
		for regName, regURL := range config.Artifactory {
			reg, err := New(regURL)
			if err != nil {
				return rm, errors.Wrap(err, "failed to create registry object")
			}
			rm[regName] = reg
		}
	}

	return rm, nil
}
