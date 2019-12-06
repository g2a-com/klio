package cmdname

import (
	"github.com/g2a-com/klio/pkg/registry"
	"strings"
)

type CmdName struct {
	Registry string
	Name     string
	original string
}

func (c CmdName) String() string {
	if len(c.original) > 0 {
		return c.original
	}
	return c.Registry + "/" + c.Name
}

func (c CmdName) DirName() string {
	return strings.Replace(c.original, "/", "__", 1)
}

func New(path string) CmdName {
	if strings.Contains(path, "__") {
		path = strings.Replace(path, "__", "/", 1)
	}
	tokens := strings.SplitN(path, "/", 2)
	if len(tokens) < 2 {
		return CmdName{
			Registry: registry.DefaultRegistryPrefix,
			Name:     tokens[0],
			original: path,
		}
	}

	return CmdName{
		Registry: tokens[0],
		Name:     tokens[1],
		original: path,
	}
}
