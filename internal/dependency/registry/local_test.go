package registry

import (
	"path"
	"reflect"
	"testing"

	"github.com/g2a-com/klio/internal/dependency"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

var fsPaths = []string{
	"/we/are/your/friends",
	"/you/ll/never/be/alone/again",
	"/well/come/on",
}

var indexFileName = "registry.yaml"

var entriesLocalUpdate = []Entry{
	{
		Name:        "docs",
		Version:     "1.0.0",
		OS:          "",
		Arch:        "",
		Annotations: map[string]string{},
	},
	{
		Name:        "docs",
		Version:     "1.2.0",
		OS:          "",
		Arch:        "",
		Annotations: map[string]string{},
	},
	{
		Name:        "docs",
		Version:     "2.1.0",
		OS:          "",
		Arch:        "",
		Annotations: map[string]string{},
	},
}

var additionalEntry = Entry{
	Name:        "docs",
	Version:     "2.1.2",
	OS:          "",
	Arch:        "",
	Annotations: map[string]string{},
}

var testIndexOne = Index{
	Entries:     entriesLocalUpdate,
	Annotations: map[string]string{},
}

var testIndexTwo = Index{
	Entries:     append(entriesLocalUpdate, additionalEntry),
	Annotations: map[string]string{},
}

func getMockFs() afero.Fs {
	fs := afero.NewMemMapFs()

	for _, p := range fsPaths {
		_ = fs.MkdirAll(p, 0x777)
	}
	indexFilePath := path.Join(fsPaths[0], indexFileName)
	file, _ := fs.Create(indexFilePath)
	_ = yaml.NewEncoder(file).Encode(testIndexOne)
	_ = file.Close()

	wrongFilePath := path.Join(fsPaths[2], indexFileName)
	wrongFile, _ := fs.Create(wrongFilePath)
	_, _ = wrongFile.Write([]byte("joke"))

	return fs
}

func changeFile(fs afero.Fs) {
	indexFilePath := path.Join(fsPaths[0], indexFileName)
	_ = fs.Remove(indexFilePath)
	file, _ := fs.Create(indexFilePath)
	_ = yaml.NewEncoder(file).Encode(testIndexTwo)
	_ = file.Close()
}

func TestLocalUpdate(t *testing.T) {
	type fields struct {
		path           string
		index          Index
		fs             afero.Fs
		currentVersion string
	}
	type want struct {
		err                bool
		errAfterChange     bool
		isThereMinorUpdate bool
		minorVersion       string
		isThereMajorUpdate bool
		majorVersion       string
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "NicelyDoneMajorAndMinor",
			fields: fields{
				path:           "/we/are/your/friends/registry.yaml",
				index:          Index{},
				fs:             getMockFs(),
				currentVersion: "1.1.0",
			},
			want: want{
				err:                false,
				errAfterChange:     false,
				isThereMinorUpdate: true,
				minorVersion:       "1.2.0",
				isThereMajorUpdate: true,
				majorVersion:       "2.1.2",
			},
		},
		{
			name: "NicelyDoneMajor",
			fields: fields{
				path:           "/we/are/your/friends/registry.yaml",
				index:          Index{},
				fs:             getMockFs(),
				currentVersion: "1.8.0",
			},
			want: want{
				err:                false,
				errAfterChange:     false,
				isThereMinorUpdate: false,
				isThereMajorUpdate: true,
				majorVersion:       "2.1.2",
			},
		},
		{
			name: "NicelyDoneMinor",
			fields: fields{
				path:           "/we/are/your/friends/registry.yaml",
				index:          Index{},
				fs:             getMockFs(),
				currentVersion: "2.0.0",
			},
			want: want{
				err:                false,
				errAfterChange:     false,
				isThereMinorUpdate: true,
				minorVersion:       "2.1.2",
				isThereMajorUpdate: false,
			},
		},
		{
			name: "NicelyDoneNone",
			fields: fields{
				path:           "/we/are/your/friends/registry.yaml",
				index:          Index{},
				fs:             getMockFs(),
				currentVersion: "2.1.2",
			},
			want: want{
				err:                false,
				errAfterChange:     false,
				isThereMinorUpdate: false,
				isThereMajorUpdate: false,
			},
		},
		{
			name: "NoSuchaAFile",
			fields: fields{
				path:  "/you/ll/never/be/alone/again/registry.yaml",
				index: Index{},
				fs:    getMockFs(),
			},
			want: want{
				err:            true,
				errAfterChange: true,
			},
		},
		{
			name: "WrongFile",
			fields: fields{
				path:  "/well/come/on/registry.yaml",
				index: Index{},
				fs:    getMockFs(),
			},
			want: want{
				err:            true,
				errAfterChange: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &local{
				path:  tt.fields.path,
				index: tt.fields.index,
				fs:    tt.fields.fs,
			}
			err := reg.Update()
			if (err != nil) != tt.want.err {
				t.Errorf("Update()[1] error = %v, wantErr %v", err, tt.want.err)
			}
			if err == nil && !reflect.DeepEqual(reg.index, testIndexOne) {
				t.Errorf("After first update got = %v, want %v", reg.index, testIndexOne)
			} else if err == nil && reflect.DeepEqual(reg.index, testIndexOne) {
				changeFile(reg.fs)
				err := reg.Update()
				if (err != nil) != tt.want.errAfterChange {
					t.Errorf("Update()[2] error = %v, wantErr %v", err, tt.want.errAfterChange)
				}
				if err == nil && !reflect.DeepEqual(reg.index, testIndexTwo) {
					t.Errorf("After second update got = %v, want %v", reg.index, testIndexTwo)
				}
			}

			dep := dependency.Dependency{
				Name:    "docs",
				Version: tt.fields.currentVersion,
			}

			nonBreaking, _ := reg.GetHighestNonBreaking(dep)
			if (nonBreaking != nil) != tt.want.isThereMinorUpdate {
				t.Errorf("GetHighestNonBreaking()[1] got = %s, want %s", nonBreaking, tt.want.minorVersion)
			}
			if (nonBreaking != nil) && (nonBreaking.Version != tt.want.minorVersion) {
				t.Errorf("GetHighestNonBreaking()[2] got = %s, want %s", nonBreaking.Version, tt.want.minorVersion)
			}
			breaking, _ := reg.GetHighestBreaking(dep)
			if (breaking != nil) != tt.want.isThereMajorUpdate {
				t.Errorf("GetHighestBreaking()[1] got = %s, want %s", breaking, tt.want.majorVersion)
			}
			if (breaking != nil) && (breaking.Version != tt.want.majorVersion) {
				t.Errorf("GetHighestBreaking()[2] got = %s, want %s", breaking.Version, tt.want.majorVersion)
			}
		})
	}
}
