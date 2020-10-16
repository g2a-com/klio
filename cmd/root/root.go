package root

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	getCommand "github.com/g2a-com/klio/cmd/get"
	"github.com/g2a-com/klio/pkg/dependency"
	"github.com/g2a-com/klio/pkg/log"
)

// NewCommand returns root command for a G2A CLI
func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "g2a",
		Short: "One tool to rule all G2A applications",
		Long: `The G2A CLI is a tool designed to get you working quickly and efficiently with
G2A services, with an emphasis on automation. It unifies and generalises all
work you need to do in order to run, build and deploy G2A services. G2A CLI is
intended to be used the same way on developer’s local machine or on CI/CD
servers like Jenkins, Bamboo or TeamCity.`,
		Version: VERSION,
	}

	// Setup flags
	verbosity := cmd.PersistentFlags().CountP("verbose", "v", "more verbose output (-vv... to further increase verbosity)")
	logLevel := cmd.PersistentFlags().String("log-level", log.GetDefaultLevel(), "set logs level: "+strings.Join(log.LevelNames, ", "))

	// Normally flags are parsed by cobra on Execute(), but we need to determine
	// logging level before executing command, so Parse() needs to be called here
	// manually. As far as I checked, this doesn't interfere with cobra parsing
	// rest of the flags later on.
	cmd.PersistentFlags().Parse(os.Args)

	// Set log level. In order to pass log level to installed subcommands we need
	// set env variable as well.
	log.SetLevel(*logLevel)
	log.IncreaseLevel(*verbosity)

	envLogLevel, ok := os.LookupEnv("G2A_CLI_LOG_LEVEL")
	if ok {
		log.SetLevel(envLogLevel)
	} else {
		_ = os.Setenv("G2A_CLI_LOG_LEVEL", log.GetLevel())
	}

	// Discover commands
	depsMgr := dependency.NewManager()
	globalCommands, err := depsMgr.GetInstalledCommands(dependency.GlobalScope)
	if err != nil {
		log.Verbose(err)
	}
	projectCommands, err := depsMgr.GetInstalledCommands(dependency.ProjectScope)
	if err != nil {
		log.Verbose(err)
	}

	// Register builtin commands
	cmd.AddCommand(getCommand.NewCommand())

	// Register external commands
	for _, dep := range projectCommands {
		loadExternalCommand(cmd, dep, false)
	}

	// Register global commands (installed under user's home directory)
	for _, path := range globalCommands {
		loadExternalCommand(cmd, path, true)
	}

	return cmd
}
