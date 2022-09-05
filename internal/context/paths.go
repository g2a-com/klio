package context

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
)

func assemblePaths(cfg CLIConfig) (Paths, error) {
	homeDir, err := getHomeDirPath()
	if err != nil {
		return Paths{}, err
	}

	projectDir, err := getProjectDir(cfg.ProjectConfigFileName)
	if err != nil {
		return Paths{}, err
	}

	return Paths{
		ProjectConfigFile: path.Join(projectDir, cfg.ProjectConfigFileName),
		ProjectInstallDir: path.Join(projectDir, cfg.InstallDirName),
		GlobalInstallDir:  path.Join(homeDir, cfg.InstallDirName),
	}, nil
}

// getHomeDirPath returns home directory of current user.
func getHomeDirPath() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("can't determine current user")
	}

	homeDir, err := filepath.EvalSymlinks(currentUser.HomeDir)
	if err != nil {
		return "", fmt.Errorf("can't fetch user directory %s", currentUser.HomeDir)
	}

	return homeDir, nil
}

func getProjectDir(configFileName string) (string, error) {
	workDir, err := getWorkDirPath()
	if err != nil {
		return "", err
	}

	for dir := workDir; dir != "/"; dir = path.Dir(dir) {
		projectConfigPath := path.Join(dir, configFileName)
		if _, err := os.Stat(projectConfigPath); err == nil {
			return dir, nil
		}
	}

	return workDir, nil
}

// getWorkDirPath returns home directory of current user.
func getWorkDirPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("can't determine working directory; %s", err)
	}

	return wd, nil
}
