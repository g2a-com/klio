package context

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

func NewCLIContext(cfg CLIConfig) CLIContext {
	return CLIContext{
		Config: cfg,
		Paths:  discoverPaths(cfg),
	}
}
