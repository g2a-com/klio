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

	workDir, err := getWorkDirPath()
	if err != nil {
		return Paths{}, err
	}

	return Paths{
		ProjectConfigFile: path.Join(workDir, cfg.ProjectConfigFileName),
		ProjectInstallDir: path.Join(workDir, cfg.InstallDirName),
		GlobalInstallDir:  path.Join(homeDir, cfg.InstallDirName),
	}, nil
}

func IsProjectConfigPresent(projectConfigFile string) bool {
	fi, err := os.Stat(projectConfigFile)
	return err == nil && !fi.IsDir()
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

// getWorkDirPath returns home directory of current user.
func getWorkDirPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("can't determine working directory; %s", err)
	}

	wdDir, err := filepath.EvalSymlinks(wd)
	if err != nil {
		return "", fmt.Errorf("can't determine fetch directory; %s", err)
	}

	return wdDir, nil
}
