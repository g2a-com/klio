package project

import (
	"os"
	"path"
	"testing"
)

func TestCreateDefaultProjectConfig(t *testing.T) {
	// prepare
	dir, err := os.MkdirTemp("", "prefix")
	if err != nil {
		t.Fatalf("can't create temporary directory: %s", err)
	}

	defaultProjectConfig := NewDefaultConfig()

	// create temporary file to test error on existing file
	existingKlioFileName := "existing-klio.yaml"
	existingKlioFileNameAbsPath := path.Join(dir, existingKlioFileName)
	defer func() { _ = os.RemoveAll(existingKlioFileNameAbsPath) }()

	_, err = os.Create(path.Join(dir, existingKlioFileName))
	if err != nil {
		t.Fatalf("can't create klio file: %s", err)
	}

	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name: "should create default config file",
			args: args{
				filePath: "klio1.yaml",
			},
			want:    defaultProjectConfig,
			wantErr: false,
		},
		{
			name: "should return error on existing file",
			args: args{
				filePath: existingKlioFileName,
			},
			want:    defaultProjectConfig,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := path.Join(dir, tt.args.filePath)
			defer func() { _ = os.Remove(filePath) }()

			_, err := CreateDefaultProjectConfig(filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDefaultProjectConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			_, err = os.Stat(filePath)
			if err != nil && !tt.wantErr {
				t.Errorf("can't find %s file", filePath)
			}
		})
	}
}
