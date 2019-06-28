package main

import (
	"fmt"
	"os"
	"time"

	"github.com/g2a-com/klio/pkg/log"

	cmd "github.com/g2a-com/klio/cmd/root"
)

var installCmd string

func main() {
	rootCmd := cmd.NewCommand()

	version := make(chan string, 1)
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(5 * time.Second)
		timeout <- true
	}()
	go cmd.CheckForNewRootVersion(version)

	err := rootCmd.Execute()

	select {
	case v := <-version:
		if v != "" {
			log.SetOutput(os.Stderr)
			log.Warnf(`there is new g2a cli version %v available - please update using: %s`, v, installCmd)
			log.SetOutput(os.Stdout)
		}
	case <-timeout:
		break
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
