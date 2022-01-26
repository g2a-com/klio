package registry

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/g2a-com/klio/internal/dependency"
	"gopkg.in/yaml.v3"
)

type wrongYaml struct {
	IsWrong int    `yaml:"wrong"`
	Why     string `yaml:"why"`
}

var (
	invalidUrlHost   = "me.not.host"
	validUrlPath     = fmt.Sprintf("/path/to/%s", indexFileName)
	invalidUrlPath   = fmt.Sprintf("/wrong/path/to/%s", indexFileName)
	wrongFileUrlPath = fmt.Sprintf("/wrong/file/%s", indexFileName)
)

func getMockArtifactory() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}

		switch r.URL.Path {
		case validUrlPath:
			out, _ := yaml.Marshal(testIndexOne)
			_, _ = w.Write(out)
		case invalidUrlPath:
			w.WriteHeader(http.StatusNotFound)
		case wrongFileUrlPath:
			wrong := wrongYaml{IsWrong: 1000, Why: "I dos no yuml"}
			out, _ := yaml.Marshal(wrong)
			_, _ = w.Write(out)
		}
	}))

	return ts
}

func TestRemoteUpdate(t *testing.T) {
	testServer := getMockArtifactory()
	defer testServer.Close()

	type fields struct {
		url            string
		client         *http.Client
		currentVersion string
	}
	type want struct {
		err                bool
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
			name: "InvalidHost",
			fields: fields{
				url:    fmt.Sprintf("%s%s", invalidUrlHost, validUrlPath),
				client: testServer.Client(),
			},
			want: want{
				err: true,
			},
		},
		{
			name: "AllGoodMajorAndMinor",
			fields: fields{
				url:            fmt.Sprintf("%s%s", testServer.URL, validUrlPath),
				client:         testServer.Client(),
				currentVersion: "1.1.0",
			},
			want: want{
				err:                false,
				isThereMinorUpdate: true,
				minorVersion:       "1.2.0",
				isThereMajorUpdate: true,
				majorVersion:       "2.1.0",
			},
		},
		{
			name: "AllGoodMajor",
			fields: fields{
				url:            fmt.Sprintf("%s%s", testServer.URL, validUrlPath),
				client:         testServer.Client(),
				currentVersion: "1.8.0",
			},
			want: want{
				err:                false,
				isThereMajorUpdate: true,
				majorVersion:       "2.1.0",
			},
		},
		{
			name: "AllGoodMinor",
			fields: fields{
				url:            fmt.Sprintf("%s%s", testServer.URL, validUrlPath),
				client:         testServer.Client(),
				currentVersion: "2.0.0",
			},
			want: want{
				err:                false,
				isThereMinorUpdate: true,
				minorVersion:       "2.1.0",
			},
		},
		{
			name: "AllGoodNone",
			fields: fields{
				url:            fmt.Sprintf("%s%s", testServer.URL, validUrlPath),
				client:         testServer.Client(),
				currentVersion: "2.1.0",
			},
			want: want{
				err: false,
			},
		},
		{
			name: "InvalidPath",
			fields: fields{
				url:    fmt.Sprintf("%s%s", testServer.URL, invalidUrlPath),
				client: testServer.Client(),
			},
			want: want{
				err: true,
			},
		},
		{
			name: "InvalidFile",
			fields: fields{
				url:    fmt.Sprintf("%s%s", testServer.URL, wrongFileUrlPath),
				client: testServer.Client(),
			},
			want: want{
				err: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &remote{
				url:    tt.fields.url,
				index:  Index{},
				client: tt.fields.client,
			}
			err := reg.Update()
			if (err != nil) != tt.want.err {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.want.err)
			}
			if err == nil && !reflect.DeepEqual(reg.index, testIndexOne) {
				t.Errorf("After update got = %v, want %v", reg.index, testIndexOne)
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
