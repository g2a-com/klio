package scope

import (
	"fmt"
	"github.com/g2a-com/klio/internal/context"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/schema"
	"os"
)

const allowedNumberOfGlobalCommands = 1

type global struct {
	dependencyManager    *dependency.Manager
	commandName          string
	GlobalInstallDir     string
	CommandFromOption    string
	CommandAsOption      string
	CommandVersionOption string
}

func NewGlobal(globalInstallDir string, commandFromOption string, commandAsOption string, commandVersionOption string) *global {
	return &global{GlobalInstallDir: globalInstallDir, CommandFromOption: commandFromOption, CommandAsOption: commandAsOption, CommandVersionOption: commandVersionOption}
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
	g.dependencyManager = dependency.NewManager(*ctx)
	g.dependencyManager.DefaultRegistry = ctx.Config.DefaultRegistry

	return nil
}

func (g *global) InstallDependencies(listOfCommands []string) error {

	if len(listOfCommands) != allowedNumberOfGlobalCommands {
		return fmt.Errorf("wrong number of commands provided; provided %d, expected %d",
			len(listOfCommands), allowedNumberOfGlobalCommands)
	}
	g.commandName = listOfCommands[0]

	dep := []schema.Dependency{
		{
			Name:     g.commandName,
			Version:  g.CommandVersionOption,
			Registry: g.CommandFromOption,
			Alias:    g.CommandAsOption,
		},
	}

	_ = installDependencies(g.dependencyManager, dep, dependency.GlobalScope)

	return nil
}

func (g *global) GetSuccessMsg() string {
	return fmt.Sprintf("command %s:%s installed successfully", g.commandName, g.CommandVersionOption)
}
