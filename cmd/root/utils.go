package root

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/g2a-com/klio/pkg/log"
)

func loadVersionFromCache(g2aDir string, label string) string {
	cacheFilePath := filepath.Join(g2aDir, "cache", "cli-version-"+label)
	cacheFile, err := os.OpenFile(cacheFilePath, os.O_RDONLY, 0644)
	defer cacheFile.Close()
	if err != nil {
		log.Spamf("failed to read version check result from cache: %s", err)
		return ""
	}

	cacheFileInfo, err := cacheFile.Stat()
	if err != nil {
		log.Spamf("failed to read version check result from cache: %s", err)
		return ""
	}

	if cacheFileInfo.ModTime().AddDate(0, 0, 1).Before(time.Now()) {
		log.Spamf("failed to read version check result from cache: %s", err)
		return ""
	}

	cacheFileContent := make([]byte, cacheFileInfo.Size())
	_, err = cacheFile.Read(cacheFileContent)
	if err != nil {
		log.Spamf("failed to read version check result from cache: %s", err)
		return ""
	}

	return strings.TrimSpace(string(cacheFileContent))
}

func saveVersionToCache(g2aDir string, label string, version string) {
	os.MkdirAll(filepath.Join(g2aDir, "cache"), 0755)

	cacheFilePath := filepath.Join(g2aDir, "cache", "cli-version-"+label)
	cacheFile, err := os.OpenFile(cacheFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	defer cacheFile.Close()
	if err != nil {
		log.Verbosef("failed to cache version check result: %s", err)
		return
	}

	_, err = cacheFile.WriteString(version)

	if err != nil {
		log.Verbosef("failed to cache version check result: %s", err)
	}
}
