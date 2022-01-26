package main

import (
	"github.com/g2a-com/klio/internal/log"
	"github.com/g2a-com/klio/pkg/cli"
)

var Version string

func main() {
	cmd := cli.CLI{
		CommandName: "klio",
		Version:     Version,
	}

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
