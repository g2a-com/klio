package cli

import (
	"errors"
	"fmt"

	"github.com/g2a-com/klio/internal/cmd/root"
	"github.com/g2a-com/klio/internal/context"
)

// TODO - add comment
type CLI struct {
	CommandName     string
	Description     string
	Version         string
	DefaultRegistry string
}

// TODO - add comment
func (cli *CLI) Execute() error {
	cfg := context.CLIConfig{
		CommandName:     cli.CommandName,
		Description:     cli.Description,
		Version:         cli.Version,
		DefaultRegistry: cli.DefaultRegistry,
	}

	if cfg.CommandName == "" {
		cfg.CommandName = "klio"
	}

	cfg.ProjectConfigFileName = fmt.Sprintf("%s.yaml", cli.CommandName)
	cfg.InstallDirName = fmt.Sprintf(".%s", cli.CommandName)

	ctx := context.NewCLIContext(cfg)
	cmd := root.NewCommand(ctx)

	if err := cmd.Execute(); err != nil {
		return errors.New("CLI execution ended with error")
	}

	return nil
}
