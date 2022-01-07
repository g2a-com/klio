package scope

import (
	"github.com/g2a-com/klio/internal/context"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/log"
	"github.com/g2a-com/klio/internal/schema"
)

type Scope interface {
	GetSuccessMsg() string
	ValidatePaths() error
	Initialize(ctx *context.CLIContext) error
	InstallDependencies(listOfCommands []string) error
}

func installDependencies(depsMgr *dependency.Manager, toInstall []schema.Dependency, scope dependency.ScopeType) []schema.Dependency {
	var installedDeps []schema.Dependency

	for _, dep := range toInstall {
		installedDep, err := depsMgr.InstallDependency(dep, scope)

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
