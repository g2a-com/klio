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
			externalCmd := exec.Command(externalCmdPath, args...)
			externalCmd.Stdin = os.Stdin
			externalCmd.Stdout = os.Stdout
			externalCmd.Stderr = os.Stderr

			version := make(chan string, 1)
			timeout := make(chan bool, 1)
			go func() {
				time.Sleep(1 * time.Second)
				timeout <- true
			}()
			go checkForNewVersion(cmdName, cmdConfig.Version, version)

			log.Debugf(`running %s "%s"`, externalCmdPath, strings.Join(args, `" "`))
			err := externalCmd.Run()

			select {
			case v := <-version:
				g := ""
				if global {
					g = "-g "
				}
				cmdGet := fmt.Sprintf("g2a get %s%s@%s", g, cmdName, v)
				log.Warnf(`there is new version %v available for %s command - please update using: %s`, v, cmdName, cmdGet)
			case <-timeout:
				break
			}

			if err != nil {
				log.Fatalf("%s", err)
				os.Exit(1)
			}
		},
		Version: cmdConfig.Version,
	}
	rootCmd.AddCommand(cmd)
}

func checkForNewVersion(cmdName string, cmdVersion string, version chan<- string) {
	if cmdVersion == "" {
		log.Spamf("version for %s not specified, unable to check for new version", cmdName)
		return
	}
	commandRegistry, err := registry.New(registry.DefaultRegistry)
	if err != nil {
		log.Spamf("failed to parse registry URL: %s", err)
		return
	}
	versions, err := commandRegistry.ListCommandVersions(cmdName)
	if err != nil {
		log.Spamf("unable to get %s command versions: %s", cmdName, err)
		return
	}
	versionConstraint, err := semver.NewConstraint(fmt.Sprintf(">%s", cmdVersion))
	if err != nil {
		log.Spamf("unable to check for new %s version: %s", cmdName, err)
		return
	}
	cmdMatchedVersion, ok := versions.MatchVersion(versionConstraint, runtime.GOOS, runtime.GOARCH)
	if ok {
		version <- strings.Replace(cmdMatchedVersion.String()[1:], fmt.Sprintf("-%s-%s", runtime.GOOS, runtime.GOARCH), "", 1)
	}
}
