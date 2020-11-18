package context

import (
	"os"
	"os/user"
	"path"
	"path/filepath"
)

func discoverPaths(cfg CLIConfig) paths {
	result := paths{}
	homeDir := getHomeDirPath()

	if homeDir != "" {
		result.GlobalInstallDir = path.Join(homeDir, cfg.InstallDirName)
	}

	result.ProjectConfigFile = findProjectConfigFile(homeDir, cfg.ProjectConfigFileName)

	if result.ProjectConfigFile != "" {
		result.ProjectInstallDir = path.Join(filepath.Dir(result.ProjectConfigFile), cfg.InstallDirName)
	}

	return result
}

// UserHomeDir returns home directory of current user
func getHomeDirPath() string {
	currentUser, err := user.Current()

	if err != nil {
		return ""
	}

	homeDir, err := filepath.EvalSymlinks(currentUser.HomeDir)

	if err != nil {
		return ""
	}

	return homeDir
}

// ProjectRootDir returns root directory (directory containing klio.yaml file)
// for a current project
func findProjectConfigFile(homeDir string, projectConfigFileName string) string {
	dir, err := os.Getwd()

	if err != nil {
		return ""
	}

	dir, err = filepath.EvalSymlinks(dir)

	if err != nil {
		return ""
	}

	for true {
		// Home directory cannot be a project directory
		if dir == homeDir {
			break
		}

		// Root directory of the filesystem cannot be a project directory
		if dir == filepath.Dir(dir) {
			break
		}

		// Project directory must contain klio.yaml file
		path := filepath.Join(dir, projectConfigFileName)
		if file, err := os.Stat(path); err == nil && !file.IsDir() {
			return path
		}

		dir = filepath.Dir(dir)
	}

	return ""
}
