package remove

import (
	"fmt"
	"strings"

	"github.com/g2a-com/klio/internal/context"
	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/log"
	"github.com/g2a-com/klio/internal/scope"
	"github.com/spf13/cobra"
)

// Options for a removeCommand command.
type options struct {
	Global bool
}

// NewCommand creates a new removeCommand command.
func NewCommand(ctx context.CLIContext) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "remove [command name]",
		Short: "Remove installed commands",
		Long:  fmt.Sprintf("Remove (%s removeCommand) will remove commands used with %s.", ctx.Config.CommandName, ctx.Config.CommandName),
		Run: func(_ *cobra.Command, args []string) {
			removeCommand(ctx, opts, args)
		},
	}

	cmd.Flags().BoolVarP(&opts.Global, "global", "g", false, "remove command installed globally")

	return cmd
}

func removeCommand(ctx context.CLIContext, opts *options, args []string) {
	var removeScope scope.Scope
	var err error

	if opts.Global {
		removeScope, err = scope.NewGlobal(&ctx)
	} else {
		removeScope, err = scope.NewLocal(&ctx, false, false)
	}
	if err != nil {
		log.Fatalf("scope initialization failed: %s", err)
	}

	var dependencies []dependency.Dependency
	switch len(args) {
	case 0:
		dependencies = removeScope.GetImplicitDependencies()
	case 1:
		dependencies = []dependency.Dependency{
			{
				Alias: args[0],
			},
		}
	default:
		log.Fatalf("max one command can be provided for removal; provided %d", len(args))
	}

	err = removeScope.RemoveDependencies(dependencies)
	if err != nil {
		log.Fatalf("removing dependencies failed: %s", err)
	}

	removedDeps := removeScope.GetRemovedDependencies()
	var formattingArray []string
	for _, d := range removedDeps {
		formattingArray = append(formattingArray, d.Alias)
	}
	log.Infof("All dependencies (%s) removed successfully", strings.Join(formattingArray, ","))
}
