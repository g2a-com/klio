package registry

import (
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"

	"github.com/g2a-com/klio/pkg/log"
	"github.com/g2a-com/klio/pkg/schema"

	"gopkg.in/yaml.v3"
)

var DefaultRegistry string
const DefaultRegistryPrefix = "g2a"

// Registry represents some commands registry hosted by Artifactory
type Registry struct {
	URL  string
	data schema.Registry
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
		defer res.Body.Close()
		buffer, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
	}

	data := schema.Registry{}
	if err := yaml.Unmarshal(buffer, &data); err != nil {
		return err
	}
	reg.data = data

	return nil
}

func (reg *Registry) FindCompatibleDependency(dep schema.Dependency) (entry *schema.RegistryEntry) {
	var result *schema.RegistryEntry
	var resultVer Version

	for idx, entry := range reg.data.Entries {
		ver := Version(entry.Version)
		if dep.Name == entry.Name && isCompatible(&entry) && ver.Match(dep.Version) && (result == nil || ver.GreaterThan(resultVer) || isMoreSpecific(&entry, result)) {
			result = &reg.data.Entries[idx]
			resultVer = ver
		}
	}

	return result
}

func isCompatible(entry *schema.RegistryEntry) bool {
	return (entry.OS == runtime.GOOS || entry.OS == "") && (entry.Arch == runtime.GOARCH || entry.Arch == "")
}

func isMoreSpecific(entry1 *schema.RegistryEntry, entry2 *schema.RegistryEntry) bool {
	return (entry1.OS != "" && entry2.OS == "") || (entry1.Arch != "" && entry2.Arch == "")
}
