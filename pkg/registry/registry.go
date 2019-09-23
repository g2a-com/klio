package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/g2a-com/klio/pkg/log"
	"github.com/g2a-com/klio/pkg/tarball"

	"github.com/otiai10/copy"
)

var DefaultRegistry string

type artifactoryFolder struct {
	Children []artifactoryChild `json:"children"`
}

type artifactoryChild struct {
	URI      string `json:"uri"`
	IsFolder bool   `json:"folder"`
}

type artifactoryFile struct {
	DownloadURI string `json:"downloadUri"`
}

// Registry represents some commands registry hosted by Artifactory
type Registry struct {
	RegistryURL string
}

// New returns new registry instance
func New(registryURL string) (*Registry, error) {
	parsedURL, err := url.Parse(registryURL)
	if err != nil {
		return nil, err
	}

	cleanedURL := &url.URL{
		Scheme: parsedURL.Scheme,
		User:   parsedURL.User,
		Host:   parsedURL.Host,
		Path:   parsedURL.Path,
	}

	registry := &Registry{
		RegistryURL: cleanedURL.String(),
	}

	return registry, nil
}

// DownloadCommand downloads command from the registry and puts it in the
// specified directory
func (reg *Registry) DownloadCommand(cmdName string, cmdVersion *CommandVersion, outputDir string) error {
	// Make request to the registry
	apiResponse, err := http.Get(reg.RegistryURL + "/" + url.PathEscape(cmdName) + "/" + cmdVersion.String() + ".tar.gz")
	if err != nil {
		return err
	}
	defer apiResponse.Body.Close()
	buffer, err := ioutil.ReadAll(apiResponse.Body)
	if err != nil {
		return err
	}
	var data artifactoryFile
	err = json.Unmarshal(buffer, &data)
	if err != nil {
		return err
	}

	tmpDir, err := ioutil.TempDir("", "g2a-cli-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	artifactResponse, err := http.Get(data.DownloadURI)
	if err != nil {
		return err
	}
	defer artifactResponse.Body.Close()

	err = tarball.Extract(artifactResponse.Body, tmpDir)
	if err != nil {
		return err
	}

	os.RemoveAll(outputDir)

	err = os.Rename(tmpDir, outputDir)
	if err != nil {
		err := copy.Copy(tmpDir, outputDir)
		if err != nil {
			return err
		}
	}

	return nil
}

// ListCommandVersions returns list of available versions of specified command
func (reg *Registry) ListCommandVersions(cmdName string) (*CommandVersionSet, error) {
	// Make request to the registry
	url := reg.RegistryURL + "/" + url.PathEscape(cmdName)
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error %d when GET %s", response.StatusCode, url)
	}
	defer response.Body.Close()
	buffer, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var data artifactoryFolder
	err = json.Unmarshal(buffer, &data)
	if err != nil {
		return nil, err
	}

	// Prepare list of versions
	var versions CommandVersionSet
	for _, child := range data.Children {
		if child.IsFolder {
			continue
		}

		versionFile := child.URI[1:]
		if !strings.HasSuffix(versionFile, fmt.Sprintf("-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)) {
			continue
		}

		versionString := strings.TrimSuffix(versionFile, ".tar.gz")
		version, err := NewCommandVersion(versionString)
		if err != nil {
			continue
		}

		versions = append(versions, *version)
	}

	return &versions, nil
}

// ListCommandVersions returns list of available versions of specified command
func (reg *Registry) ListRootVersions() (*CommandVersionSet, error) {
	// Make request to the registry
	response, err := http.Get(reg.RegistryURL + "/")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	buffer, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var data artifactoryFolder
	err = json.Unmarshal(buffer, &data)
	if err != nil {
		return nil, err
	}

	// Prepare list of versions
	var versions CommandVersionSet
	for _, child := range data.Children {
		if child.IsFolder {
			continue
		}

		versionFile := child.URI[1:]

		if !strings.HasSuffix(versionFile, fmt.Sprintf("-%s-%s", runtime.GOOS, runtime.GOARCH)) {
			continue
		}

		version, err := NewCommandVersion(versionFile)
		if err != nil {
			log.Debugf("found g2a cli version file with invalid name: '%s'", versionFile)
			continue
		}

		versions = append(versions, *version)
	}

	return &versions, nil
}
