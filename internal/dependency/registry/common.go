package registry

import (
	"runtime"

	"github.com/g2a-com/klio/internal/config"
	"github.com/g2a-com/klio/internal/dependency"
)

type Registry interface {
	Update() error
	GetHighestBreaking(dep dependency.Dependency) (*Entry, error)
	GetHighestNonBreaking(dep dependency.Dependency) (*Entry, error)
	GetExactMatch(dep dependency.Dependency) (*Entry, error)
}

type Index struct {
	Meta        config.Metadata   `yaml:"-"`
	APIVersion  string            `yaml:"apiVersion,omitempty"`
	Kind        string            `yaml:"kind,omitempty"`
	Annotations map[string]string `yaml:"annotations"`
	Entries     []Entry           `yaml:"entries"`
}

type Entry struct {
	Name        string            `yaml:"name"`
	Version     string            `yaml:"version"`
	OS          string            `json:"os"`
	Arch        string            `json:"arch"`
	Annotations map[string]string `yaml:"annotations"`
	URL         string            `yaml:"url"`
	Checksum    string            `yaml:"checksum"`
}

func findHighestMatching(registryEntries []Entry, currentDependency dependency.Dependency, constraintFunction func(version Version) (string, error)) (*Entry, error) {
	var result *Entry

	constraint, err := constraintFunction(Version(currentDependency.Version))
	if err != nil {
		return nil, err
	}

	for idx, entry := range registryEntries {
		ver := Version(entry.Version)
		if currentDependency.Name == entry.Name && isCompatible(entry) && ver.Match(constraint) && (result == nil || ver.GreaterThan(Version(result.Version)) || isMoreSpecific(entry, *result)) {
			result = &registryEntries[idx]
		}
	}

	return result, nil
}

func isCompatible(entry Entry) bool {
	return (entry.OS == runtime.GOOS || entry.OS == "") && (entry.Arch == runtime.GOARCH || entry.Arch == "")
}

func isMoreSpecific(entry1 Entry, entry2 Entry) bool {
	return (entry1.OS != "" && entry2.OS == "") || (entry1.Arch != "" && entry2.Arch == "")
}
