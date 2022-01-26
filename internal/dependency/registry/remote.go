package registry

import (
	"fmt"
	"io"
	"net/http"
	"reflect"

	"gopkg.in/yaml.v3"

	"github.com/g2a-com/klio/internal/dependency"
	"github.com/g2a-com/klio/internal/log"
	"gopkg.in/yaml.v3"
)

// remote represents registry hosted on http server.
type remote struct {
	url   string
	index Index
	fs    afero.Fs
}

// NewRemote returns new registry instance hosted on http server.
func NewRemote(registryUrl string) Registry {
	registry := &remote{
		url: registryUrl,
		fs:  afero.NewOsFs(),
	}

	return registry
}

func (reg *remote) Update() error {
	log.Spamf("Loading registry: %s", reg.url)

	var buffer []byte

	res, err := http.Get(reg.url)
	if err != nil {
		return err
	}
	if res.StatusCode >= http.StatusMultipleChoices { // 300
		return fmt.Errorf("artifactory returned response: %s", res.Status)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)
	buffer, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	data := Index{}
	if err := yaml.Unmarshal(buffer, &data); err != nil {
		return err
	}
	reg.index = data

	return nil
}

func (reg *remote) GetExactMatch(dep dependency.Dependency) (*Entry, error) {
	return findHighestMatching(reg.index.Entries, dep, getExactMatch)
}

func (reg *remote) GetHighestBreaking(dep dependency.Dependency) (*Entry, error) {
	return findHighestMatching(reg.index.Entries, dep, getMajorConstraints)
}

func (reg *remote) GetHighestNonBreaking(dep dependency.Dependency) (*Entry, error) {
	return findHighestMatching(reg.index.Entries, dep, getMinorAndPatchConstraints)
}
