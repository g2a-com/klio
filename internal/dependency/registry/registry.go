package registry

import (
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"

	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/log"

	"gopkg.in/yaml.v3"
)

// Registry represents some commands registry hosted by Artifactory
type Registry struct {
	URL        string
	configFile Config
}

// New returns new registry instance
func New(registryURL string) *Registry {
	registry := &Registry{
		URL: registryURL,
	}

	return registry
}

func (reg *Registry) Update() error {
	log.Spamf("Loading registry: %s", reg.URL)

	var err error
	var buffer []byte

	if strings.HasPrefix(reg.URL, "file:///") {
		path := reg.URL[7:] // remove "file://"
		buffer, err = ioutil.ReadFile(path)
		if err != nil {
			return err
		}
	} else {
		res, err := http.Get(reg.URL)
		if err != nil {
			return err
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(res.Body)
		buffer, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
	}

	data := Config{}
	if err := yaml.Unmarshal(buffer, &data); err != nil {
		return err
	}
	reg.configFile = data

	return nil
}

func (reg *Registry) FindCompatibleDependency(dep dependency.Dependency) (entry *Entry) {
	var result *Entry
	var resultVer Version

	for idx, entry := range reg.configFile.Entries {
		ver := Version(entry.Version)
		if dep.Name == entry.Name && isCompatible(&entry) && ver.Match(dep.Version) && (result == nil || ver.GreaterThan(resultVer) || isMoreSpecific(&entry, result)) {
			result = &reg.configFile.Entries[idx]
			resultVer = ver
		}
	}

	return result
}

func isCompatible(entry *Entry) bool {
	return (entry.OS == runtime.GOOS || entry.OS == "") && (entry.Arch == runtime.GOARCH || entry.Arch == "")
}

func isMoreSpecific(entry1 *Entry, entry2 *Entry) bool {
	return (entry1.OS != "" && entry2.OS == "") || (entry1.Arch != "" && entry2.Arch == "")
}
