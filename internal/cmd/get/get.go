package get

import (
	"fmt"

	"github.com/g2a-com/klio/internal/context"
	"github.com/g2a-com/klio/internal/log"
	"github.com/g2a-com/klio/internal/scope"
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
	var getScope scope.Scope

	if opts.Global {
		getScope = scope.NewGlobal(ctx.Paths.GlobalInstallDir, opts.From, opts.As, opts.Version)
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
	err = getScope.InstallDependencies(args)
	if err != nil {
		log.Fatalf("installing dependencies failed: %s", err)
	}
	log.Info(getScope.GetSuccessMsg())
}
