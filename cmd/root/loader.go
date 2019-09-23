package root

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/g2a-com/klio/pkg/cmdname"
	"github.com/g2a-com/klio/pkg/config"
	"github.com/g2a-com/klio/pkg/discover"
	"github.com/g2a-com/klio/pkg/log"
	"github.com/g2a-com/klio/pkg/registry"
)

func loadExternalCommand(rootCmd *cobra.Command, commandConfigPath string, global bool) {
	cmdDir := filepath.Dir(commandConfigPath)

	cmdName := cmdname.New(filepath.Base(filepath.Dir(commandConfigPath)))
	if cmd, _, _ := rootCmd.Find([]string{cmdName.Name}); cmd != rootCmd {
		log.Debugf("cannot register already registered command '%s' provided by '%s'", cmdName, cmdDir)
		return
	}

	cmdConfig, err := config.LoadCommandConfig(commandConfigPath)
	if err != nil {
		log.Warnf("cannot load command: %s", err)
		return
	}

	cmd := &cobra.Command{
		Use:                cmdName.Name,
		Short:              cmdConfig.Description,
		Long:               "",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			externalCmdPath := filepath.Join(cmdDir, cmdConfig.BinPath)
			var externalCmd *exec.Cmd
			if runtime.GOOS == "windows" {
				args = append([]string{"/c", externalCmdPath}, args...)
				externalCmdPath = "cmd"
			}
			externalCmd = exec.Command(externalCmdPath, args...)
			externalCmd.Stdin = os.Stdin
			externalCmd.Stdout = os.Stdout
			externalCmd.Stderr = os.Stderr

			version := make(chan string, 1)
			timeout := make(chan bool, 1)
			go func() {
				time.Sleep(5 * time.Second)
				timeout <- true
			}()
			go checkForNewVersion(filepath.Dir(cmdConfig.Meta.Path), cmdName, cmdConfig.Version, version)

			_ = os.Setenv("G2A_CLI_GLOBAL_COMMAND", strconv.FormatBool(global))

			log.Debugf(`running %s "%s"`, externalCmdPath, strings.Join(args, `" "`))
			err := externalCmd.Run()

			select {
			case v := <-version:
				if v != "" {
					g := ""
					if global {
						g = "-g "
					}
					cmdGet := fmt.Sprintf("g2a get %s%s@%s", g, cmdName, v)
					log.SetOutput(os.Stderr)
					log.Warnf(`there is new version %v available for %s command - please update using: %s`, v, cmdName, cmdGet)
					log.SetOutput(os.Stdout)
				}
			case <-timeout:
				break
			}

			if err != nil {
				switch e := err.(type) {
				case *exec.ExitError:
					os.Exit(e.ExitCode())
				default:
					log.Fatal(err)
				}
			}
		},
		Version: cmdConfig.Version,
	}
	rootCmd.AddCommand(cmd)
}

func checkForNewVersion(cmdDir string, cmdName cmdname.CmdName, cmdVersion string, version chan<- string) {
	result := loadVersionFromCache("command-" + cmdName.DirName())

	if cmdVersion == "" {
		log.Spamf("version for %s not specified, unable to check for new version", cmdName.String())
		version <- ""
		return
	}
	versionConstraint, err := semver.NewConstraint(fmt.Sprintf(">%s", cmdVersion))
	if err != nil {
		log.Spamf("unable to check for new %s version: %s", cmdName.String(), err)
		version <- ""
		return
	}

	if result == "" {
		commandRegistry, err := loadRegistry(cmdName)
		if err != nil {
			log.Spam(err.Error())
			version <- ""
			return
		}

		versions, err := commandRegistry.ListCommandVersions(cmdName.Name)
		if err != nil {
			log.Spamf("unable to get %s command versions: %s", cmdName.String(), err)
			version <- ""
			return
		}
		cmdMatchedVersion, ok := versions.MatchVersion(versionConstraint, runtime.GOOS, runtime.GOARCH)
		if !ok {
			log.Spamf("no new versions of '%s' command", cmdName.String())
			result = cmdVersion
		} else {
			result = strings.Replace(cmdMatchedVersion.String()[1:], fmt.Sprintf("-%s-%s", runtime.GOOS, runtime.GOARCH), "", 1)
		}

		saveVersionToCache("command-"+cmdName.DirName(), result)
	}

	if ver, err := semver.NewVersion(result); err == nil && versionConstraint.Check(ver) {
		version <- result
	} else {
		version <- ""
	}
}

func loadRegistry(cmdName cmdname.CmdName) (*registry.Registry, error) {
	if cmdName.Registry == "default" {
		var err error
		reg, err := registry.New(registry.DefaultRegistry)
		return reg, err
	}

	baseDir, ok := discover.ProjectRootDir()
	if !ok {
		return nil, errors.New("not in project directory - aborting version check")
	}
	projectConfig, err := config.LoadProjectConfig(filepath.Join(baseDir, "g2a.yaml"))
	if err != nil {
		return nil, errors.Wrap(err, "error reading project config")
	}
	regMap, err := registry.NewRegistriesMap(projectConfig.CLI.Registries)
	if err != nil {
		return nil, err
	}
	reg, ok := regMap[cmdName.Registry]
	if !ok {
		return nil, fmt.Errorf("command registry not found for %s", cmdName.String())
	}
	return reg, nil
}
