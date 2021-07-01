package registry

import "github.com/g2a-com/klio/internal/configfile"

type Config struct {
	Meta        configfile.Metadata `yaml:"-"`
	APIVersion  string              `yaml:"apiVersion,omitempty"`
	Kind        string              `yaml:"kind,omitempty"`
	Annotations map[string]string   `yaml:"annotations"`
	Entries     []Entry             `yaml:"entries"`
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

type EntryVersion struct {
	Number string `yaml:"number" json:"number"`
	OS     string `yaml:"os" json:"os"`
	Arch   string `yaml:"arch" json:"arch"`
}
