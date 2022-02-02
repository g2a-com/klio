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

func installDependencies(depsMgr *manager.Manager, toInstall []dependency.Dependency, installDir string) []dependency.Dependency {
	var installedDeps []dependency.Dependency

	for _, dep := range toInstall {
		installedDep, err := depsMgr.InstallDependency(dep, installDir)
		if err != nil {
			log.LogfAndExit(log.FatalLevel, "Failed to install %s@%s: %s", dep.Name, dep.Version, err)
		}

		if installedDep.Alias == "" {
			log.Infof("Installed %s@%s from %s", installedDep.Name, installedDep.Version, installedDep.Registry)
		} else {
			log.Infof("Installed %s@%s from %s as %s", installedDep.Name, installedDep.Version, installedDep.Registry, installedDep.Alias)
		}

		installedDeps = append(installedDeps, *installedDep)
	}

	return installedDeps
}
