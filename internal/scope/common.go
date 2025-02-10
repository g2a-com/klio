package scope

import (
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/dependency/manager"
	"github.com/g2a-com/klio/internal/log"
)

type Scope interface {
	GetImplicitDependencies() []dependency.Dependency
	InstallDependencies([]dependency.Dependency) ([]dependency.Dependency, []dependency.DependenciesIndexEntry, error)
	GetInstalledDependencies() []dependency.Dependency
	RemoveDependencies([]dependency.Dependency) error
	GetRemovedDependencies() []dependency.Dependency
}

func installDependencies(depsMgr *manager.Manager, toInstall []dependency.Dependency, installDir string) ([]dependency.Dependency, []dependency.DependenciesIndexEntry, error) {
	var installedDeps []dependency.Dependency
	var installedDepsIndex []dependency.DependenciesIndexEntry

	for _, dep := range toInstall {
		depIndexEntry, err := depsMgr.InstallDependency(&dep, installDir)
		if err != nil {
			return nil, nil, err
		}

		if dep.Alias == "" {
			log.Infof("Installed %s@%s from %s", dep.Name, dep.Version, dep.Registry)
		} else {
			log.Infof("Installed %s@%s from %s as %s", dep.Name, dep.Version, dep.Registry, dep.Alias)
		}

		installedDeps = append(installedDeps, dep)
		installedDepsIndex = append(installedDepsIndex, *depIndexEntry)
	}

	return installedDeps, installedDepsIndex, nil
}

func removeDependencies(depsMgr *manager.Manager, toRemove []dependency.Dependency, installDir string) []dependency.Dependency {
	var removedDeps []dependency.Dependency

	for _, dep := range toRemove {

		if err := depsMgr.RemoveDependency(&dep, installDir); err != nil {
			log.Fatalf("Failed to remove %s: %s", dep.Alias, err)
		}

		log.Infof("Removed %s", dep.Alias)

		removedDeps = append(removedDeps, dep)
	}

	return removedDeps
}
