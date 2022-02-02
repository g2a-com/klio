package registry

import (
	"strings"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"

	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/log"
)

// local represents registry hosted locally.
type local struct {
	path  string
	index Index
	fs    afero.Fs
}

// NewLocal returns new registry instance hosted locally.
func NewLocal(registryPath string) Registry {
	registry := &local{
		path: registryPath,
		fs:   afero.NewOsFs(),
	}

	return registry
}

func (reg *local) Update() error {
	log.Spamf("Loading registry: %s", reg.path)

	var buffer []byte

	path := strings.TrimPrefix(reg.path, "file://")
	file, err := reg.fs.Open(path)
	if err != nil {
		return err
	}
	buffer, err = afero.ReadAll(file)
	if err != nil {
		return err
	}

	var i Index
	if err := yaml.Unmarshal(buffer, &i); err != nil {
		return err
	}
	reg.index = i

	return nil
}

func (reg *local) GetExactMatch(dep dependency.Dependency) (*Entry, error) {
	return findHighestMatching(reg.index.Entries, dep, getExactMatch)
}

func (reg *local) GetHighestBreaking(dep dependency.Dependency) (*Entry, error) {
	return findHighestMatching(reg.index.Entries, dep, getMajorConstraints)
}

func (reg *local) GetHighestNonBreaking(dep dependency.Dependency) (*Entry, error) {
	return findHighestMatching(reg.index.Entries, dep, getMinorAndPatchConstraints)
}
