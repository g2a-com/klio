package get

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"stash.code.g2a.com/CLI/core/pkg/discover"
	"stash.code.g2a.com/CLI/core/pkg/registry"

	"github.com/Masterminds/semver"
	"stash.code.g2a.com/CLI/core/pkg/log"

	"github.com/spf13/cobra"
)

const defaultRegistry = "https://artifactory.code.g2a.com/artifactory/api/storage/g2a-cli-local"

// Options for a get command
type options struct {
	Global   bool
	Registry string
}

// NewCommand creates a new get command
func NewCommand() *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:     "get [command name]",
		Short:   "Install new commands",
		Long:    "Get (g2a get) will install command to use with G2A CLI.",
		Run:     opts.run,
		Version: "test",
	}

	cmd.Flags().BoolVarP(&opts.Global, "global", "g", false, "install command globally")
	cmd.Flags().StringVar(&opts.Registry, "registry", defaultRegistry, "change address to the registry")

	return cmd
}

func (opts *options) run(cmd *cobra.Command, args []string) {
	// Find directory for installing packages
	var baseDir string
	var ok bool

	if opts.Global {
		baseDir, ok = discover.UserHomeDir()
		if !ok {
			log.Fatal("cannot find user's home directory")
		}
	} else {
		baseDir, ok = discover.ProjectRootDir()
		if !ok {
			log.Fatal(`packages can be installed locally only under project directory, use "--global" option`)
		}
	}

	dir := filepath.Join(baseDir, filepath.FromSlash(".g2a/cli-commands"))

	registry, err := registry.New(opts.Registry)
	if err != nil {
		log.Errorf("failed to parse registry URL: %s", err)
		os.Exit(1)
	}

	// Download packages
	exitCode := 0

	for _, arg := range args {
		argParts := strings.SplitN(arg, "@", 2)
		cmdName := argParts[0]
		versionRange := "*"

		if len(argParts) >= 2 {
			versionRange = argParts[1]
		}

		versionConstraint, err := semver.NewConstraint(versionRange)
		if err != nil {
			log.Errorf("invalid version range %s", arg)
			exitCode = 1
			continue
		}

		versions, err := registry.ListCommandVersions(cmdName)
		if err != nil {
			log.Error(err)
			exitCode = 1
			continue
		}
		log.Spamf("found following versions for %s: %s", arg, versions.String())

		cmdVersion, ok := versions.MatchVersion(versionConstraint, runtime.GOOS, runtime.GOARCH)
		if !ok {
			log.Errorf("no matching version found for %s", arg)
			exitCode = 1
			continue
		}
		log.Spamf("found matching version for %s: %s", arg, versions)

		err = registry.DownloadCommand(cmdName, cmdVersion, dir)
		if err != nil {
			exitCode = 1
			log.Errorf("failed to download %s@%s: %s", cmdName, cmdVersion.Version, err)
		} else {
			log.Infof("downloaded %s@%s", cmdName, cmdVersion.Version)
		}
	}

	os.Exit(exitCode)
}
