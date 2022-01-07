package cli

import (
	"errors"
	"fmt"

	"github.com/g2a-com/klio/internal/cmd/root"
	"github.com/g2a-com/klio/internal/context"
)

// CLI defines a custom-made cli
// TODO: launch the validation.
type CLI struct {
	// CommandName invokes your custom CLI together with any potential subcommands.
	// Treat it as the parent command.
	// E.g. for CommandName _klio_ the command line execution will start with `klio`.
	CommandName string `validate:"required"`
	// Description of your CLI - describe its purpose, when to use it.
	Description string
	// Version of your CLI release.
	// There is no restriction on version naming, although a version tag must be provided.
	// We recommend to use semver convention while tagging versions.
	Version string `validate:"required"`
	// DefaultRegistry url will be used anytime a user fails to provide explicit registry in `get` subcommand.
	DefaultRegistry string `validate:"url"`
}

// Execute the base command to validate its configuration and launch subcommand specified in command line.
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

	ctx, err := context.Initialize(cfg)
	if err != nil {
		return fmt.Errorf("initialization error: %s", err)
	}
	cmd := root.NewCommand(ctx)

	if err := cmd.Execute(); err != nil {
		return errors.New("CLI execution ended with error")
	}

	return nil
}
