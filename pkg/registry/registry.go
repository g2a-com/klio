package registry

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/g2a-com/klio/pkg/log"
	"github.com/g2a-com/klio/pkg/tarball"
)

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
func (reg *Registry) DownloadCommand(cmdName string, cmdVersion *CommandVersion, outputPath string) error {
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

	outputDir := filepath.Join(outputPath, cmdName)
	os.RemoveAll(outputDir)
	os.Rename(tmpDir, outputDir)

	return nil
}

// ListCommandVersions returns list of available versions of specified command
func (reg *Registry) ListCommandVersions(cmdName string) (*CommandVersionSet, error) {
	// Make request to the registry
	response, err := http.Get(reg.RegistryURL + "/" + url.PathEscape(cmdName))
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
		if !strings.HasSuffix(versionFile, ".tar.gz") {
			log.Debugf("found command version file with invalid extension: %s/%s", cmdName, versionFile)
			continue
		}

		versionString := strings.TrimSuffix(versionFile, ".tar.gz")
		version, err := NewCommandVersion(versionString)
		if err != nil {
			log.Debugf("found command version file with invalid name: '%s/%s'", cmdName, versionFile)
			continue
		}

		versions = append(versions, *version)
	}

	return &versions, nil
}
