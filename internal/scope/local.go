package scope

import (
	"fmt"
	"os"
	"os/user"
	"path"

	"github.com/g2a-com/klio/internal/context"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/dependency/manager"
	"github.com/g2a-com/klio/internal/project"
	"github.com/spf13/afero"
)

const (
	standardDirPermission = 0o755
)

type local struct {
	projectConfig     *project.Config
	dependencyManager *manager.Manager
	installedDeps     []dependency.Dependency
	removedDeps       []dependency.Dependency
	os                afero.Fs
	projectConfigFile string
	installDir        string
	noInit            bool
	noSave            bool
}

func NewLocal(projectConfigFile string, projectInstallDir string, noInit bool, noSave bool) *local {
	return &local{projectConfigFile: projectConfigFile, installDir: projectInstallDir, noInit: noInit, noSave: noSave, os: afero.NewOsFs()}
}

func (l *local) ValidatePaths() error {
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("can't determine current user")
	}

	projectDir := path.Dir(l.projectConfigFile)

	if projectDir == currentUser.HomeDir {
		return fmt.Errorf("home directory cannot be a project directory")
	}

	if projectDir == path.Dir(projectDir) {
		return fmt.Errorf("root directory of the filesystem cannot be a project directory")
	}

	return nil
}

func (l *local) Initialize(ctx *context.CLIContext) error {

	// look for config file
	configFile, configFileErr := l.os.Stat(l.projectConfigFile)
	if !l.noInit {
		// make sure install dir exists
		_ = l.os.MkdirAll(l.installDir, standardDirPermission)
		var err error
		// make sure if config file exists
		if os.IsNotExist(configFileErr) {
			_, err = project.CreateDefaultProjectConfig(l.projectConfigFile)
			if err != nil {
				return err
			}
		} else if err == nil && configFile.IsDir() {
			return fmt.Errorf("can't create config file; path collision with a directory %s", l.projectConfigFile)
		}
	} else if os.IsNotExist(configFileErr) {
		return fmt.Errorf("%s not found; make sure it exists before running command with \"--no-init\"", ctx.Config.ProjectConfigFileName)
	}

	// initialize dependency manager
	l.dependencyManager = manager.NewManager()
	l.dependencyManager.DefaultRegistry = ctx.Config.DefaultRegistry

	// load project config
	var err error
	l.projectConfig, err = project.LoadProjectConfig(ctx.Paths.ProjectConfigFile)
	if err != nil {
		return err
	}

	return nil
}

func (l *local) GetImplicitDependencies() []dependency.Dependency {
	return l.projectConfig.Dependencies
}

func (l *local) InstallDependencies(listOfCommands []dependency.Dependency) error {
	if len(listOfCommands) == 0 {
		return fmt.Errorf("no dependencies provided for the project")
	}

	installedDeps, err := installDependencies(l.dependencyManager, listOfCommands, l.installDir)
	if err != nil {
		return err
	}
	l.installedDeps = installedDeps

	if !l.noSave {
		for _, installedDep := range l.installedDeps {
			var idx int
			var projectDep dependency.Dependency
			for idx, projectDep = range l.projectConfig.Dependencies {
				if projectDep.Alias == installedDep.Alias {
					l.projectConfig.Dependencies[idx] = installedDep
					break
				}
			}
			if idx != len(l.projectConfig.Dependencies) {
			} else {
				l.projectConfig.Dependencies = append(l.projectConfig.Dependencies, installedDep)
			}
		}

		l.projectConfig.DefaultRegistry = l.dependencyManager.DefaultRegistry

		if err := project.SaveProjectConfig(l.projectConfig); err != nil {
			return fmt.Errorf("unable to update dependencies in the %s file: %s", l.projectConfigFile, err)
		}
	}

	return nil
}

func (l *local) GetInstalledDependencies() []dependency.Dependency {
	return l.installedDeps
}

func (l *local) RemoveDependencies(listOfCommands []dependency.Dependency) error {
	if len(listOfCommands) == 0 {
		return fmt.Errorf("no dependencies provided for the project")
	}

	l.removedDeps = removeDependencies(l.dependencyManager, listOfCommands, l.installDir)

	return nil
}

func (l *local) GetRemovedDependencies() []dependency.Dependency {
	return l.removedDeps
}
