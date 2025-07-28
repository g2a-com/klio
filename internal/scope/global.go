package scope

import (
	"fmt"
	"os"

	"github.com/g2a-com/klio/internal/context"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/dependency/manager"
	"github.com/spf13/afero"
)

type global struct {
	os                afero.Fs
	dependencyManager *manager.Manager
	installedDeps     []dependency.Dependency
	removedDeps       []dependency.Dependency
	installDir        string
}

func NewGlobal(ctx *context.CLIContext) (*global, error) {
	g := &global{
		installDir: ctx.Paths.GlobalInstallDir,
		os:         afero.NewOsFs(),
	}
	if err := g.validatePaths(); err != nil {
		return nil, err
	}
	if err := g.initialize(ctx); err != nil {
		return nil, err
	}
	return g, nil
}

func (g *global) validatePaths() error {
	if _, err := g.os.Stat(g.installDir); os.IsNotExist(err) {
		// make sure install dir exists
		err = g.os.MkdirAll(g.installDir, standardDirPermission)
		if err != nil {
			return fmt.Errorf("global dir initialization failed with error: %s", err.Error())
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (g *global) initialize(ctx *context.CLIContext) error {
	// initialize dependency manager
	g.dependencyManager = manager.NewManager()
	g.dependencyManager.DefaultRegistry = ctx.Config.DefaultRegistry
	installedCommands := g.dependencyManager.GetInstalledCommands(ctx.Paths)
	for _, command := range installedCommands {
		if ctx.Paths.IsGlobal(command.Path) {
			g.installedDeps = append(g.installedDeps, command.ToDependency())
		}
	}

	return nil
}

func (g *global) GetImplicitDependencies() []dependency.Dependency {
	return g.installedDeps
}

func (g *global) InstallDependencies(listOfCommands []dependency.Dependency) ([]dependency.Dependency, []dependency.DependenciesIndexEntry, error) {
	installedDeps, installedDepsEntries, err := installDependencies(g.dependencyManager, listOfCommands, g.installDir)
	if err != nil {
		return nil, nil, err
	}
	g.installedDeps = installedDeps

	return installedDeps, installedDepsEntries, nil
}

func (g *global) GetInstalledDependencies() []dependency.Dependency {
	return g.installedDeps
}

func (g *global) RemoveDependencies(listOfCommands []dependency.Dependency) error {
	if len(listOfCommands) == 0 {
		return fmt.Errorf("no dependencies provided for removal")
	}

	g.removedDeps = removeDependencies(g.dependencyManager, listOfCommands, g.installDir)

	return nil
}

func (g *global) GetRemovedDependencies() []dependency.Dependency {
	return g.removedDeps
}
