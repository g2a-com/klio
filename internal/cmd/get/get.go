package get

import (
	"fmt"
	"os"
	"path"

	"github.com/g2a-com/klio/internal/context"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/log"
	"github.com/g2a-com/klio/internal/schema"
	"github.com/spf13/cobra"
)

// Options for a get command.
type options struct {
	Global  bool
	NoSave  bool
	From    string
	As      string
	Version string
	NoInit  bool
}

// NewCommand creates a new get command.
func NewCommand(ctx context.CLIContext) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "get [command name]",
		Short: "Install new commands",
		Long:  fmt.Sprintf("Get (%s get) will install command to use with %s.", ctx.Config.CommandName, ctx.Config.CommandName),
		Run: func(cmd *cobra.Command, args []string) {
			run(ctx, opts, cmd, args)
		},
	}

	cmd.Flags().BoolVarP(&opts.Global, "global", "g", false, "install command globally")
	cmd.Flags().BoolVar(&opts.NoSave, "no-save", false, "prevent saving to dependencies")
	cmd.Flags().StringVar(&opts.From, "from", "", "address of the registry")
	cmd.Flags().StringVar(&opts.As, "as", "", "changes name under which dependency is installed")
	cmd.Flags().BoolVar(&opts.NoInit, "no-init", false, "prevent creating config file if not exist")
	cmd.Flags().StringVar(&opts.Version, "version", "*", "version range of the dependency")

	return cmd
}

func run(ctx context.CLIContext, opts *options, _ *cobra.Command, args []string) {
	// Find directory for installing packages
	var projectConfig *schema.ProjectConfig
	var err error
	var scope dependency.ScopeType
	var installedDeps []schema.Dependency

	depsMgr := dependency.NewManager(ctx)
	depsMgr.DefaultRegistry = ctx.Config.DefaultRegistry

	if opts.Global {
		scope = dependency.GlobalScope
		if ctx.Paths.GlobalInstallDir == "" {
			log.Fatal("Cannot init global install directory")
		}
	} else {
		scope = dependency.ProjectScope

		if !opts.NoInit && ctx.Paths.ProjectInstallDir == "" {
			ctx, err = initialiseProjectInCurrentDir(ctx)
			if err != nil {
				log.Fatalf("Failed to initialise project in the current dir: %s", err)
			}
			depsMgr = dependency.NewManager(ctx)
		}

		if ctx.Paths.ProjectInstallDir == "" {
			log.Fatal(`Packages can be installed locally only under project directory, use "--global"`)
		}
		projectConfig, err = schema.LoadProjectConfig(ctx.Paths.ProjectConfigFile)
		if err != nil {
			log.Fatal(err)
		}

		if projectConfig.DefaultRegistry != "" {
			depsMgr.DefaultRegistry = projectConfig.DefaultRegistry
		}
	}

	var toInstall []schema.Dependency

	if len(args) == 0 && !opts.Global {
		toInstall = projectConfig.Dependencies
	} else {
		toInstall = []schema.Dependency{
			schema.Dependency{
				Name:     args[0],
				Version:  opts.Version,
				Registry: opts.From,
				Alias:    opts.As,
			},
		}
	}

	for _, dep := range toInstall {
		installedDep, err := depsMgr.InstallDependency(dep, scope)

		if err != nil {
			log.LogfAndExit(log.FatalLevel, "Failed to install %s@%s: %s", dep.Name, dep.Version, err)
		} else {
			if installedDep.Alias == "" {
				log.Infof("Installed %s@%s from %s", installedDep.Name, installedDep.Version, installedDep.Registry)
			} else {
				log.Infof("Installed %s@%s from %s as %s", installedDep.Name, installedDep.Version, installedDep.Registry, installedDep.Alias)
			}
		}

		installedDeps = append(installedDeps, *installedDep)
	}

	if !opts.Global && !opts.NoSave {
		for _, installedDep := range installedDeps {
			var idx int
			for idx = 0; idx < len(projectConfig.Dependencies); idx++ {
				if projectConfig.Dependencies[idx].Alias == installedDep.Alias {
					break
				}
			}
			if idx != len(projectConfig.Dependencies) {
				projectConfig.Dependencies[idx] = installedDep
			} else {
				projectConfig.Dependencies = append(projectConfig.Dependencies, installedDep)
			}
		}

		projectConfig.DefaultRegistry = depsMgr.DefaultRegistry

		if err := schema.SaveProjectConfig(projectConfig); err != nil {
			log.Errorf("Unable to update dependencies in the %s file: %s", ctx.Config.ProjectConfigFileName, err)
		} else {
			log.Infof("Updated dependencies in the %s file", ctx.Config.ProjectConfigFileName)
		}
	}
}

// initialiseProjectInCurrentDir creates default klio.yaml file in current directory and update context.
func initialiseProjectInCurrentDir(ctx context.CLIContext) (context.CLIContext, error) {
	// get current directory
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		return ctx, err
	}

	return initialiseProject(ctx, currentWorkingDirectory)
}

// initialiseProject creates default klio.yaml file in given directory and update context.
func initialiseProject(ctx context.CLIContext, dirPath string) (context.CLIContext, error) {
	// update context
	ctx.Paths.ProjectInstallDir = path.Join(dirPath, ctx.Config.InstallDirName)
	ctx.Paths.ProjectConfigFile = path.Join(dirPath, ctx.Config.ProjectConfigFileName)

	// make sure install dir exists
	err := os.MkdirAll(ctx.Paths.ProjectInstallDir, 0o755)
	if err != nil {
		return ctx, err
	}

	// create and save default klio config
	_, err = schema.CreateDefaultProjectConfig(ctx.Paths.ProjectConfigFile)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}
