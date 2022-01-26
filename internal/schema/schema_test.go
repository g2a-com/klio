package schema

import (
	"testing"

	"github.com/g2a-com/klio/internal/config"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestNewDefaultProjectConfig(t *testing.T) {
	tests := []struct {
		name string
		want *ProjectConfig
	}{
		{
			name: "should return default project config values",
			want: &ProjectConfig{
				Meta:            config.Metadata{},
				DefaultRegistry: "",
				Dependencies:    nil,
				yaml: &yaml.Node{
					Kind: yaml.MappingNode,
					Tag:  "!!map",
					Content: []*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Tag:   "!!str",
							Value: "dependencies",
						},
						{
							Kind: yaml.MappingNode,
							Tag:  "!!map",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDefaultProjectConfig()
			assert.EqualValues(t, tt.want, got)
		})
	}
}
