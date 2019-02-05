package subcommand

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/g2a-com/klio/pkg/log"
)

func Init(cmd *cobra.Command) {
	rootCmd := &cobra.Command{Use: "g2a"}
	rootCmd.PersistentFlags().CountP("verbose", "v", "more verbose output (-vv... to further increase verbosity)")
	rootCmd.PersistentFlags().String("log-level", log.GetDefaultLevel(), "set logs level: "+strings.Join(log.LevelNames, ", "))
	rootCmd.AddCommand(cmd)
}

func Execute(cmd *cobra.Command) {
	// There is no other way to modify args passed to cmd.Execute()...
	os.Args = append([]string{os.Args[0], cmd.Name()}, os.Args[1:]...)

	err := cmd.Execute()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
