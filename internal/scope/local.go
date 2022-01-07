package scope

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/g2a-com/klio/internal/context"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/schema"
)

const (
	allowedNumberOfLocalCommands = 0
	standardDirPermission        = 0o755
)

type local struct {
	projectConfig     *schema.ProjectConfig
	dependencyManager *dependency.Manager
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
	if !l.NoInit && !ctx.ProjectConfigExists {
		// make sure install dir exists
		_ = os.MkdirAll(l.ProjectInstallDir, standardDirPermission)

		// make sure if config file exists
		if configFile, err := os.Stat(l.ProjectConfigFile); os.IsNotExist(err) {
			_, err = schema.CreateDefaultProjectConfig(l.ProjectConfigFile)
			if err != nil {
				return err
			}
			ctx.ProjectConfigExists = true
		} else if err == nil && configFile.IsDir() {
			return fmt.Errorf("can't create config file; path collision with a directory %s", l.ProjectConfigFile)
		}
	}

	// if project was not setup by user nor automatically, throw error
	if !ctx.ProjectConfigExists {
		return fmt.Errorf(`packages can be installed locally only under project directory, use "--global"`)
	}

	// initialize dependency manager
	l.dependencyManager = dependency.NewManager(*ctx)
	l.dependencyManager.DefaultRegistry = ctx.Config.DefaultRegistry

	// load project config
	var err error
	l.projectConfig, err = schema.LoadProjectConfig(ctx.Paths.ProjectConfigFile)
	if err != nil {
		return err
	}

	return nil
}

func (l *local) InstallDependencies(listOfCommands []string) error {
	if len(listOfCommands) != allowedNumberOfLocalCommands {
		return fmt.Errorf("wrong number of commands provided; provided %d, expected %d",
			len(listOfCommands), allowedNumberOfLocalCommands)
	}

	installedDeps := installDependencies(l.dependencyManager, l.projectConfig.Dependencies, dependency.GlobalScope)

	if !l.NoSave {
		for _, installedDep := range installedDeps {
			var idx int
			var projectDep schema.Dependency
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

func (l *local) GetSuccessMsg() string {
	var formattingArray []string

	if len(l.projectConfig.Dependencies) == 0 {
		return "no dependencies installed, ensure that the config file is correct"
	}

	for _, d := range l.projectConfig.Dependencies {
		formattingArray = append(formattingArray, fmt.Sprintf("%s:%s", d.Alias, d.Version))
	}
	return fmt.Sprintf("all project dependencies (%s) installed successfully", strings.Join(formattingArray, ","))
}
