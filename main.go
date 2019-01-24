package main

import (
	"fmt"
	"os"

	cmd "stash.code.g2a.com/CLI/core/cmd/root"
)

func main() {
	rootCmd := cmd.NewCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
