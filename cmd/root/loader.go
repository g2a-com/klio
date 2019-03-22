package root

import (
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"stash.code.g2a.com/cli/common/pkg/config"
	"github.com/g2a-com/klio/pkg/log"
)

func loadExternalCommand(rootCmd *cobra.Command, commandConfigPath string) {
	cmdDir := path.Dir(commandConfigPath)

	cmdName := path.Base(path.Dir(commandConfigPath))
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
			externalCmdPath := path.Join(cmdDir, cmdConfig.BinPath)
			externalCmd := exec.Command(externalCmdPath, args...)
			externalCmd.Stdin = os.Stdin
			externalCmd.Stdout = os.Stdout
			externalCmd.Stderr = os.Stderr

			log.Debugf(`running %s "%s"`, externalCmdPath, strings.Join(args, `" "`))
			err := externalCmd.Run()
			if err != nil {
				os.Exit(1)
			}
		},
		Version: cmdConfig.Version,
	}
	rootCmd.AddCommand(cmd)
}
