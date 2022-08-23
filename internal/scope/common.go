package scope

import (
	"github.com/g2a-com/klio/internal/context"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/dependency/manager"
	"github.com/g2a-com/klio/internal/log"
)

type Scope interface {
	ValidatePaths() error
	Initialize(*context.CLIContext) error
	GetImplicitDependencies() []dependency.Dependency
	InstallDependencies([]dependency.Dependency) error
	GetInstalledDependencies() []dependency.Dependency
}

func installDependencies(depsMgr *manager.Manager, toInstall []dependency.Dependency, installDir string) ([]dependency.Dependency, error) {
	var installedDeps []dependency.Dependency

	for _, dep := range toInstall {

		if err := depsMgr.InstallDependency(&dep, installDir); err != nil {
			return nil, err
		}

		if dep.Alias == "" {
			log.Infof("Installed %s@%s from %s", dep.Name, dep.Version, dep.Registry)
		} else {
			log.Infof("Installed %s@%s from %s as %s", dep.Name, dep.Version, dep.Registry, dep.Alias)
		}

		installedDeps = append(installedDeps, dep)
	}

	return installedDeps, nil
}
