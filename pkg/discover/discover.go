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

// LocalCommandPaths returns list of paths for custom commands installed locally
// in the repository
func LocalCommandPaths() []string {
	projectDir, ok := ProjectRootDir()

	if !ok {
		return []string{}
	}

	globPattern := filepath.Join(projectDir, filepath.FromSlash(".g2a/cli-commands/*/command.so"))
	paths, err := filepath.Glob(globPattern)

	if err != nil {
		return []string{}
	}

	return paths
}

// UserCommandPaths returns list of paths for custom commands installed
// under the user's home directory
func UserCommandPaths() []string {
	homeDir, ok := UserHomeDir()

	if !ok {
		return []string{}
	}

	globPattern := filepath.Join(homeDir, filepath.FromSlash(".g2a/cli-commands/*/command.so"))
	paths, err := filepath.Glob(globPattern)

	if err != nil {
		return []string{}
	}

	return paths
}
