package root

import (
	"os"

	"path"

	"github.com/spf13/cobra"
	"github.com/g2a-com/klio/pkg/log"
	"github.com/g2a-com/klio/pkg/runner"
	"github.com/g2a-com/klio/pkg/util"
)

type packageJson struct {
	Description string `json:"description"`
	Version     string `json:"version"`
	BinPath     string `json:"binPath" validate:"required,file"`
}

func loadExternalCommand(rootCmd *cobra.Command, packageJsonPath string) {
	cmdDir := path.Dir(packageJsonPath)

	cmdName := path.Base(path.Dir(packageJsonPath))
	if cmd, _, _ := rootCmd.Find([]string{cmdName}); cmd != rootCmd {
		log.Debugf("cannot register already registered command '%s' provided by '%s'", cmdName, cmdDir)
		return
	}

	cmdConfig := &packageJson{}
	if err := util.LoadConfigFile(cmdConfig, packageJsonPath); err != nil {
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
			externalCmd := runner.NewCommand(externalCmdPath, args...)
			externalCmd.DecorateOutput = false
			err := externalCmd.Run()
			if err != nil {
				os.Exit(1)
			}
		},
		Version: cmdConfig.Version,
	}
	rootCmd.AddCommand(cmd)
}
