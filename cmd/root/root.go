package root

import (
	"fmt"
	"github.com/Masterminds/semver"
	"os"
	"runtime"
	"github.com/g2a-com/klio/pkg/registry"
	"strings"

	"github.com/spf13/cobra"
	getCommand "github.com/g2a-com/klio/cmd/get"
	"github.com/g2a-com/klio/pkg/discover"
	"github.com/g2a-com/klio/pkg/log"
)

const VERSION = "2.2.3"

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
	os.Setenv("G2A_CLI_LOG_LEVEL", log.GetLevel())

	// Register builtin commands
	cmd.AddCommand(getCommand.NewCommand())

	// Register local commands (installed under project directory)
	for _, path := range discover.LocalCommandPaths() {
		loadExternalCommand(cmd, path, false)
	}

	// Register global commands (installed under user's home directory)
	for _, path := range discover.UserCommandPaths() {
		loadExternalCommand(cmd, path, true)
	}

	return cmd
}

func CheckForNewRootVersion(version chan<- string) {
	commandRegistry, err := registry.New(registry.DefaultRegistry)
	if err != nil {
		log.Spamf("failed to parse registry URL: %s", err)
		return
	}
	versions, err := commandRegistry.ListRootVersions()
	if err != nil {
		log.Spamf("unable to get g2a cli versions: %s", err)
		return
	}
	versionConstraint, err := semver.NewConstraint(fmt.Sprintf(">%s", VERSION))
	if err != nil {
		log.Spamf("unable to check for new g2a cli version: %s", err)
		return
	}
	cmdMatchedVersion, ok := versions.MatchVersion(versionConstraint, runtime.GOOS, runtime.GOARCH)
	if ok {
		version <- strings.Replace(cmdMatchedVersion.String()[1:], fmt.Sprintf("-%s-%s", runtime.GOOS, runtime.GOARCH), "", 1)
	}
}
