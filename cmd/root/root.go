package root

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"stash.code.g2a.com/CLI/core/pkg/discover"
	"stash.code.g2a.com/CLI/core/pkg/log"
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
		Version: "2.0.0-alpha.1",
	}

	// Setup flags
	verbosity := cmd.PersistentFlags().CountP("verbose", "v", "more verbose output (-vv... to further increase verbosity)")
	logLevel := cmd.PersistentFlags().String("log-level", "info", "set logs level: "+strings.Join(log.LevelNames, ", "))

	// Normally flags are parsed by cobra on Execute(), but we need to determine
	// logging level before executing command, so Parse() needs to be called here
	// manually. As far as I checked, this doesn't interfere with cobra parsing
	// rest of the flags later on.
	cmd.PersistentFlags().Parse(os.Args)
	log.SetLevel(*logLevel, log.InfoLevel)
	log.IncreaseLevel(*verbosity)

	// Register local commands (installed under project directory)
	for _, path := range discover.LocalCommandPaths() {
		loadExternalCommand(cmd, path)
	}

	// Register global commands (installed under user's home directory)
	for _, path := range discover.UserCommandPaths() {
		loadExternalCommand(cmd, path)
	}

	return cmd
}
