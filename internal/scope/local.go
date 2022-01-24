package scope

import (
	"fmt"
	"github.com/g2a-com/klio/internal/context"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/dependency/manager"
	"github.com/g2a-com/klio/internal/schema"
	"os"
	"os/user"
	"path"
)

const (
	standardDirPermission = 0o755
)

type local struct {
	projectConfig     *schema.ProjectConfig
	dependencyManager *manager.Manager
	installedDeps     []dependency.Dependency
	ProjectConfigFile string
	ProjectInstallDir string
	NoInit            bool
	NoSave            bool
}

func NewLocal(projectConfigFile string, projectInstallDir string, noInit bool, noSave bool) *local {
	return &local{ProjectConfigFile: projectConfigFile, ProjectInstallDir: projectInstallDir, NoInit: noInit, NoSave: noSave}
}

func (l *local) ValidatePaths() error {
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("can't determine current user")
	}

	projectDir := path.Dir(l.ProjectConfigFile)

	if projectDir == currentUser.HomeDir {
		return fmt.Errorf("home directory cannot be a project directory")
	}

	if projectDir == path.Dir(projectDir) {
		return fmt.Errorf("root directory of the filesystem cannot be a project directory")
	}

	return nil
}

func (l *local) Initialize(ctx *context.CLIContext) error {
	if !l.NoInit {
		// make sure install dir exists
		_ = os.MkdirAll(l.ProjectInstallDir, standardDirPermission)

		// make sure if config file exists
		if configFile, err := os.Stat(l.ProjectConfigFile); os.IsNotExist(err) {
			_, err = schema.CreateDefaultProjectConfig(l.ProjectConfigFile)
			if err != nil {
				return err
			}
		} else if err == nil && configFile.IsDir() {
			return fmt.Errorf("can't create config file; path collision with a directory %s", l.ProjectConfigFile)
		}
	}

	// initialize dependency manager
	l.dependencyManager = manager.NewManager(*ctx)
	l.dependencyManager.DefaultRegistry = ctx.Config.DefaultRegistry

	// load project config
	var err error
	l.projectConfig, err = schema.LoadProjectConfig(ctx.Paths.ProjectConfigFile)
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

	l.installedDeps = installDependencies(l.dependencyManager, listOfCommands, manager.GlobalScope)

	if !l.NoSave {
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

		if err := schema.SaveProjectConfig(l.projectConfig); err != nil {
			return fmt.Errorf("unable to update dependencies in the %s file: %s", l.ProjectConfigFile, err)
		}
	}

	return nil
}

func (l *local) GetInstalledDependencies() []dependency.Dependency {
	return l.installedDeps
}
