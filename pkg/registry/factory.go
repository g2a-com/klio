package registry

import (
	"github.com/pkg/errors"

	"github.com/g2a-com/klio/pkg/config"
)

func NewRegistriesMap(config *config.CmdRegistries) (map[string]*Registry, error) {
	rm := make(map[string]*Registry)
	{
		defaultRegistry, err := New(DefaultRegistry)
		if err != nil {
			return rm, errors.Wrap(err, "failed to create registry object")
		}
		rm[DefaultRegistryPrefix] = defaultRegistry
	}

	for regName, registry := range *config {
		if registry.RegistryType.Artifactory != nil {
			reg, err := New(*registry.Artifactory)
			if err != nil {
				return rm, errors.Wrap(err, "failed to create registry object")
			}
			rm[regName] = reg
		}
	}

	return rm, nil
}
