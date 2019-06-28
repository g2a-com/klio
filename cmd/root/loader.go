package root

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/g2a-com/klio/pkg/registry"

	"github.com/spf13/cobra"
	"stash.code.g2a.com/cli/common/pkg/config"
	"github.com/g2a-com/klio/pkg/log"
)

func loadExternalCommand(rootCmd *cobra.Command, commandConfigPath string, global bool) {
	cmdDir := filepath.Dir(commandConfigPath)

	cmdName := filepath.Base(filepath.Dir(commandConfigPath))
	if cmd, _, _ := rootCmd.Find([]string{cmdName}); cmd != rootCmd {
		log.Debugf("cannot register already registered command '%s' provided by '%s'", cmdName, cmdDir)
		return
	}

	cmdConfig, err := config.LoadCommandConfig(commandConfigPath)
	if err != nil {
		log.Warnf("cannot load command: %s", err)
		return
	}

	cmd := &cobra.Command{
		Use:                cmdName,
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

func checkForNewVersion(cmdDir string, cmdName string, cmdVersion string, version chan<- string) {
	g2aDir := filepath.Dir(filepath.Dir(cmdDir))
	result := loadVersionFromCache(g2aDir, "command-"+cmdName)

	if result == "" {
		if cmdVersion == "" {
			log.Spamf("version for %s not specified, unable to check for new version", cmdName)
			version <- ""
			return
		}
		commandRegistry, err := registry.New(registry.DefaultRegistry)
		if err != nil {
			log.Spamf("failed to parse registry URL: %s", err)
			version <- ""
			return
		}
		versions, err := commandRegistry.ListCommandVersions(cmdName)
		if err != nil {
			log.Spamf("unable to get %s command versions: %s", cmdName, err)
			version <- ""
			return
		}
		versionConstraint, err := semver.NewConstraint(fmt.Sprintf(">%s", cmdVersion))
		if err != nil {
			log.Spamf("unable to check for new %s version: %s", cmdName, err)
			version <- ""
			return
		}
		cmdMatchedVersion, ok := versions.MatchVersion(versionConstraint, runtime.GOOS, runtime.GOARCH)
		if !ok {
			log.Spamf("no new versions of '%s' command", cmdName)
			result = cmdVersion
		} else {
			result = strings.Replace(cmdMatchedVersion.String()[1:], fmt.Sprintf("-%s-%s", runtime.GOOS, runtime.GOARCH), "", 1)
		}

		saveVersionToCache(g2aDir, "command-"+cmdName, result)
	}

	if result != cmdVersion {
		version <- result
	} else {
		version <- ""
	}
}
