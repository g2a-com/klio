package discover

import (
	"os"
	"os/user"
	"path/filepath"
)

// UserHomeDir returns home directory of current user
func UserHomeDir() (string, bool) {
	currentUser, err := user.Current()

	if err != nil {
		return "", false
	}

	homeDir, err := filepath.EvalSymlinks(currentUser.HomeDir)

	if err != nil {
		return "", false
	}

	return homeDir, true
}

// ProjectRootDir returns root directory (directory containing g2a.yaml file)
// for a current project
func ProjectRootDir() (string, bool) {
	dir, err := os.Getwd()

	if err != nil {
		return "", false
	}

	dir, err = filepath.EvalSymlinks(dir)

	if err != nil {
		return "", false
	}

	userHomeDir, _ := UserHomeDir()

	for true {
		// Home directory cannot be a project directory
		if dir == userHomeDir {
			return "", false
		}

		// Root directory of the filesystem cannot be a project directory
		if dir == filepath.Dir(dir) {
			return "", false
		}

		// Project directory must contain g2a.yaml file
		if file, err := os.Stat(filepath.Join(dir, "g2a.yaml")); err == nil && !file.IsDir() {
			break
		}

		dir = filepath.Dir(dir)
	}

	return dir, true
}

func ProjectKlioDir() (string, bool) {
	if projectDir, ok := ProjectRootDir(); ok {
		return filepath.Join(projectDir, ".g2a"), true
	} else {
		return "", false
	}
}

func GlobalKlioDir() (string, bool) {
	if projectDir, ok := UserHomeDir(); ok {
		return filepath.Join(projectDir, ".g2a"), true
	} else {
		return "", false
	}
}
