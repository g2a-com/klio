package get

import (
	"fmt"
	"strings"

	"github.com/g2a-com/klio/internal/context"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/log"
	"github.com/g2a-com/klio/internal/scope"
	"github.com/spf13/cobra"
)

// Options for a getCommand command.
type options struct {
	Global  bool
	NoSave  bool
	From    string
	As      string
	Version string
	NoInit  bool
}

// NewCommand creates a new getCommand command.
func NewCommand(ctx context.CLIContext) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "get [command name]",
		Short: "Install new commands",
		Long:  fmt.Sprintf("Get (%s getCommand) will install command to use with %s.", ctx.Config.CommandName, ctx.Config.CommandName),
		Run: func(_ *cobra.Command, args []string) {
			getCommand(ctx, opts, args)
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

func getCommand(ctx context.CLIContext, opts *options, args []string) {
	var getScope scope.Scope

	if opts.Global {
		getScope = scope.NewGlobal(ctx.Paths.GlobalInstallDir)
	} else {
		getScope = scope.NewLocal(ctx.Paths.ProjectConfigFile, ctx.Paths.ProjectInstallDir, opts.NoInit, opts.NoSave)
	}

	err := getScope.ValidatePaths()
	if err != nil {
		log.Fatalf("validation of paths failed: %s", err)
	}
	err = getScope.Initialize(&ctx)
	if err != nil {
		log.Fatalf("scope initialization failed: %s", err)
	}

	var dependencies []dependency.Dependency
	switch len(args) {
	case 0:
		dependencies = getScope.GetImplicitDependencies()
	case 1:
		dependencies = []dependency.Dependency{
			{
				Name:     args[0],
				Registry: opts.From,
				Version:  opts.Version,
				Alias:    opts.As,
			},
		}
	default:
		log.Fatalf("max one command can be provided for install; provided %d", len(args))
	}

	err = getScope.InstallDependencies(dependencies)
	if err != nil {
		log.Fatalf("installing dependencies failed: %s", err)
	}

	installedDeps := getScope.GetInstalledDependencies()
	var formattingArray []string
	for _, d := range installedDeps {
		formattingArray = append(formattingArray, fmt.Sprintf("%s:%s", d.Alias, d.Version))
	}
	log.Infof("All dependencies (%s) installed successfully", strings.Join(formattingArray, ","))
}
