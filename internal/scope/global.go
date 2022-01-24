package scope

import (
	"fmt"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/dependency/manager"
	"os"

	"github.com/g2a-com/klio/internal/context"
)

const allowedNumberOfGlobalCommands = 1

type global struct {
	dependencyManager    *manager.Manager
	commandName          string
	installedDeps        []dependency.Dependency
	GlobalInstallDir     string
	CommandVersionOption string
}

func NewGlobal(globalInstallDir string) *global {
	return &global{GlobalInstallDir: globalInstallDir}
}

func (g *global) ValidatePaths() error {
	if _, err := os.Stat(g.GlobalInstallDir); os.IsNotExist(err) {
		return fmt.Errorf("global install dir does not exists")
	} else if err != nil {
		return err
	}
	return nil
}

func (g *global) Initialize(ctx *context.CLIContext) error {
	// initialize dependency manager
	g.dependencyManager = manager.NewManager(*ctx)
	g.dependencyManager.DefaultRegistry = ctx.Config.DefaultRegistry

	return nil
}

func (g *global) GetImplicitDependencies() []dependency.Dependency {
	return []dependency.Dependency{}
}

func (g *global) InstallDependencies(listOfCommands []dependency.Dependency) error {
	if len(listOfCommands) != allowedNumberOfGlobalCommands {
		return fmt.Errorf("wrong number of commands provided; provided %d, expected %d",
			len(listOfCommands), allowedNumberOfGlobalCommands)
	}

	dep := listOfCommands

	g.installedDeps = installDependencies(g.dependencyManager, dep, manager.GlobalScope)

	return nil
}

func (g *global) GetInstalledDependencies() []dependency.Dependency {
	return g.installedDeps
}
