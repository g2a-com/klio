package get

import (
	"path/filepath"

	"github.com/g2a-com/klio/pkg/dependency"
	"github.com/g2a-com/klio/pkg/discover"
	"github.com/g2a-com/klio/pkg/log"
	"github.com/g2a-com/klio/pkg/schema"

	"github.com/spf13/cobra"
)

var defaultRegistry string

// Options for a get command
type options struct {
	Global  bool
	NoSave  bool
	From    string
	As      string
	Version string
}

// NewCommand creates a new get command
func NewCommand() *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "get [command name]",
		Short: "Install new commands",
		Long:  "Get (g2a get) will install command to use with G2A CLI.",
		Run:   opts.run,
	}

	cmd.Flags().BoolVarP(&opts.Global, "global", "g", false, "install command globally")
	cmd.Flags().BoolVar(&opts.NoSave, "no-save", false, "prevent saving to dependencies")
	cmd.Flags().StringVar(&opts.From, "from", "", "address of the registry")
	cmd.Flags().StringVar(&opts.As, "as", "", "changes name under which dependency is installed")
	cmd.Flags().StringVar(&opts.Version, "version", "*", "version range of the dependency")

	return cmd
}

func (opts *options) run(cmd *cobra.Command, args []string) {
	// Find directory for installing packages
	var baseDir string
	var projectConfig *schema.ProjectConfig
	var ok bool
	var err error
	var scope dependency.ScopeType
	var installedDeps []schema.Dependency
	var registry string

	depsMgr := dependency.NewManager()
	depsMgr.DefaultRegistry = defaultRegistry

	if opts.Global {
		scope = dependency.GlobalScope
		baseDir, ok = discover.UserHomeDir()
		if !ok {
			log.Fatal("Cannot find user's home directory")
		}
	} else {
		scope = dependency.ProjectScope
		baseDir, ok = discover.ProjectRootDir()
		if !ok {
			log.Fatal(`Packages can be installed locally only under project directory, use "--global" option`)
		}
		projectConfig, err = schema.LoadProjectConfig(filepath.Join(baseDir, "g2a.yaml"))
		if err != nil {
			log.Fatal(err)
		}

		if registry != "" {
			registry = opts.From
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
			log.Errorf("Unable to update dependencies in the g2a.yaml file: %s", err)
		} else {
			log.Infof("Updated dependencies in the g2a.yaml file")
		}
	}
}
