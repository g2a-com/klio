package root

import (
	"os"
	"strings"

	getCommand "github.com/g2a-com/klio/internal/cmd/get"
	removeCommand "github.com/g2a-com/klio/internal/cmd/remove"
	"github.com/g2a-com/klio/internal/context"
	"github.com/g2a-com/klio/internal/dependency/manager"
	"github.com/g2a-com/klio/internal/log"
	"github.com/spf13/cobra"
)

// NewCommand returns root command for a klio.
func NewCommand(ctx context.CLIContext) *cobra.Command {
	rootCommand := &cobra.Command{
		Use:     ctx.Config.CommandName,
		Long:    ctx.Config.Description,
		Version: ctx.Config.Version,
	}

	// Setup flag
	verbosity := rootCommand.PersistentFlags().CountP("verbose", "v", "more verbose output (-vv... to further increase verbosity)")
	logLevel := rootCommand.PersistentFlags().String("log-level", log.GetDefaultLevel(), "set logs level: "+strings.Join(log.LevelNames, ", "))

	// Normally flags are parsed by cobra on Execute(), but we need to determine
	// logging level before executing command, so Parse() needs to be called here
	// manually. As far as I checked, this doesn't interfere with cobra parsing
	// rest of the flags later on.
	_ = rootCommand.PersistentFlags().Parse(os.Args)

	// Set log level. In order to pass log level to installed subcommands we need
	// set env variable as well.
	log.SetLevel(*logLevel)
	log.IncreaseLevel(*verbosity)

	envLogLevel, ok := os.LookupEnv("KLIO_LOG_LEVEL")
	if ok {
		log.SetLevel(envLogLevel)
	} else {
		_ = os.Setenv("KLIO_LOG_LEVEL", log.GetLevel())
	}

	// Discover commands
	commands := manager.NewManager().GetInstalledCommands(ctx.Paths)

	// Register builtin commands
	rootCommand.AddCommand(getCommand.NewCommand(ctx))
	rootCommand.AddCommand(removeCommand.NewCommand(ctx))

	// Register external commands
	for _, dep := range commands {
		loadExternalCommand(ctx, rootCommand, dep)
	}

	return rootCommand
}
