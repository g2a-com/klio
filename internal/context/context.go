package context

import "strings"

type CLIContext struct {
	Config CLIConfig
	Paths  Paths
}

type CLIConfig struct {
	CommandName           string
	Description           string
	Version               string
	ProjectConfigFileName string
	InstallDirName        string
	DefaultRegistry       string
}

type Paths struct {
	ProjectConfigFile string
	ProjectInstallDir string
	GlobalInstallDir  string
}

func Initialize(cfg CLIConfig) (CLIContext, error) {
	paths, err := assemblePaths(cfg)
	if err != nil {
		return CLIContext{}, err
	}

	return CLIContext{
		Config: cfg,
		Paths:  paths,
	}, nil
}

func (p Paths) IsGlobal(testedPath string) bool {
	return strings.HasPrefix(testedPath, p.GlobalInstallDir)
}

func (p Paths) IsProject(testedPath string) bool {
	return strings.HasPrefix(testedPath, p.ProjectInstallDir)
}
