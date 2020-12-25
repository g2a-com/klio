package get

import (
	"errors"
	"github.com/g2a-com/klio/internal/context"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func Test_initialiseProjectInCurrentDir(t *testing.T) {
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("can't get current directory: %s", err)
	}

	projetConfigFileName := "test-config-name.yaml"
	installDirName := "test-dir"

	type args struct {
		ctx context.CLIContext
	}
	tests := []struct {
		name    string
		args    args
		want    context.CLIContext
		wantErr bool
	}{
		{
			name: "should initialise default klio config file and update context paths",
			args: args{
				ctx: struct {
					Config context.CLIConfig
					Paths  context.Paths
				}{
					Config: context.CLIConfig{
						ProjectConfigFileName: projetConfigFileName,
						InstallDirName:        installDirName,
					},
					Paths: struct {
						ProjectConfigFile string
						ProjectInstallDir string
						GlobalInstallDir  string
					}{}},
			},
			want: context.CLIContext{
				Config: context.CLIConfig{
					ProjectConfigFileName: projetConfigFileName,
					InstallDirName:        installDirName,
				},
				Paths: context.Paths{
					ProjectConfigFile: filepath.Join(currentWorkingDirectory, projetConfigFileName),
					ProjectInstallDir: filepath.Join(currentWorkingDirectory, installDirName),
					GlobalInstallDir:  "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := initialiseProjectInCurrentDir(tt.args.ctx)
			defer os.RemoveAll(got.Paths.GlobalInstallDir)
			defer os.RemoveAll(got.Paths.ProjectConfigFile)

			if (err != nil) != tt.wantErr {
				t.Errorf("initialiseProjectInCurrentDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func Test_run(t *testing.T) {
	type args struct {
		ctx  context.CLIContext
		opts *options
		cmd  *cobra.Command
		args []string
	}
	ctx := context.CLIContext{
		Config: context.CLIConfig{},
		Paths:  context.Paths{
			ProjectConfigFile: path.Join(os.TempDir(), "config.yaml"),
			ProjectInstallDir: os.TempDir(),
			GlobalInstallDir:  os.TempDir(),
		},
	}
	rootCmdWithGet := cobra.Command{}
	emptyCmd := cobra.Command{}

	getCmd := NewCommand(ctx)
	rootCmdWithGet.AddCommand(getCmd)

	tests := []struct {
		name  string
		args  args
		error error
	}{
		{
			name: "should return error 'Cannot get already registered command 'get''",
			args: args{
				ctx: ctx,
				opts: &options{
					As: "get",
				},
				cmd:  &rootCmdWithGet,
				args: nil,
			},
			error: errors.New("Cannot get already registered command 'get'"),
		},
		{
			name: "should allow to register command 'get' since it is not registered",
			args: args{
				ctx: ctx,
				opts: &options{
					As: "get",
				},
				cmd:  &emptyCmd,
				args: nil,
			},
			error: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(tt.args.ctx, tt.args.opts, tt.args.cmd, tt.args.args)
			assert.EqualValues(t, err, tt.error)
		})
	}
}
