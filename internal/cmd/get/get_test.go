package get

import (
	"os"
	"path"
	"testing"

	"github.com/g2a-com/klio/internal/context"
	"github.com/stretchr/testify/assert"
)

const (
	projectConfigFileName = "test-config-name.yaml"
	installDirName        = "test-dir"
)

func TestInitialiseProjectInCurrentDir(t *testing.T) {
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("can't get current directory: %s", err)
	}

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
						ProjectConfigFileName: projectConfigFileName,
						InstallDirName:        installDirName,
					},
					Paths: struct {
						ProjectConfigFile string
						ProjectInstallDir string
						GlobalInstallDir  string
					}{},
				},
			},
			want: context.CLIContext{
				Config: context.CLIConfig{
					ProjectConfigFileName: projectConfigFileName,
					InstallDirName:        installDirName,
				},
				Paths: context.Paths{
					ProjectConfigFile: path.Join(currentWorkingDirectory, projectConfigFileName),
					ProjectInstallDir: path.Join(currentWorkingDirectory, installDirName),
					GlobalInstallDir:  "",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := initialiseProjectInCurrentDir(tt.args.ctx)
			defer func(path string) {
				_ = os.RemoveAll(path)
			}(got.Paths.GlobalInstallDir)
			defer func(path string) {
				_ = os.RemoveAll(path)
			}(got.Paths.ProjectConfigFile)

			if (err != nil) != tt.wantErr {
				t.Errorf("initialiseProjectInCurrentDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.EqualValues(t, tt.want, got)
		})
	}
}
