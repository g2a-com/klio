package main

import (
	"fmt"
	"os"

	cmd "github.com/g2a-com/klio/cmd/root"
)

func main() {
	rootCmd := cmd.NewCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
