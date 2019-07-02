package get

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/g2a-com/klio/pkg/config"
	"github.com/g2a-com/klio/pkg/discover"
	"github.com/g2a-com/klio/pkg/log"
	"github.com/g2a-com/klio/pkg/registry"

	"github.com/Masterminds/semver"
	"github.com/spf13/cobra"
)

// Options for a get command
type options struct {
	Global   bool
	Registry string
	NoSave   bool
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
	cmd.Flags().StringVar(&opts.Registry, "registry", registry.DefaultRegistry, "change address to the registry")
	cmd.Flags().BoolVar(&opts.NoSave, "no-save", false, "prevent saving to dependencies")

	return cmd
}

func (opts *options) run(cmd *cobra.Command, args []string) {
	// Find directory for installing packages
	var baseDir string
	var projectConfig *config.ProjectConfig
	var ok bool
	var err error

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
		projectConfig, err = config.LoadProjectConfig(filepath.Join(baseDir, "g2a.yaml"))
		if err != nil {
			log.Fatal(err)
		}
	}

	var commandsToInstall map[string]string

	if len(args) == 0 && !opts.Global {
		opts.NoSave = true // don't save g2a.yaml when installing from it
		commandsToInstall = projectConfig.CLI.Commands
	} else {
		commandsToInstall = map[string]string{}
		for _, arg := range args {
			argParts := strings.SplitN(arg, "@", 2)
			if len(argParts) == 1 {
				commandsToInstall[argParts[0]] = "*"
			} else {
				commandsToInstall[argParts[0]] = argParts[1]
			}
		}
	}

	dir := filepath.Join(baseDir, filepath.FromSlash(".g2a/cli-commands"))
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Errorf("unable to create directory: %s due to %s", dir, err)
			os.Exit(1)
		}
	}

	commandRegistry, err := registry.New(opts.Registry)
	if err != nil {
		log.Errorf("failed to parse registry URL: %s", err)
		os.Exit(1)
	}

	// Download packages
	exitCode := 0
	installedCommands := map[string]string{}

	for cmdName, versionRange := range commandsToInstall {
		versionConstraint, err := semver.NewConstraint(versionRange)
		if err != nil {
			log.Errorf("invalid version range %s@%s", cmdName, versionRange)
			exitCode = 1
			continue
		}

		versions, err := commandRegistry.ListCommandVersions(cmdName)
		if err != nil {
			log.Error(err)
			exitCode = 1
			continue
		}
		log.Spamf("found following versions for %s@%s: %s", cmdName, versionRange, versions.String())

		cmdVersion, ok := versions.MatchVersion(versionConstraint, runtime.GOOS, runtime.GOARCH)
		if !ok {
			log.Errorf("no matching version found for %s@%s", cmdName, versionRange)
			exitCode = 1
			continue
		}
		log.Spamf("found matching version for %s@%s: %s", cmdName, versionRange, versions)

		err = commandRegistry.DownloadCommand(cmdName, cmdVersion, dir)
		if err != nil {
			log.Errorf("failed to download %s@%s: %s", cmdName, cmdVersion.Version, err)
			exitCode = 1
			continue
		} else {
			log.Infof("downloaded %s@%s", cmdName, cmdVersion.Version)
		}

		installedCommands[cmdName] = cmdVersion.Version.String()
	}

	if !opts.Global && !opts.NoSave {
		for c, v := range installedCommands {
			projectConfig.CLI.Commands[c] = v
		}
		err := config.SaveProjectConfig(projectConfig)
		if err != nil {
			log.Errorf("unable to update commands list in the g2a.yaml file")
		}
		log.Infof("updated commands list in the g2a.yaml file")
	}

	os.Exit(exitCode)
}
