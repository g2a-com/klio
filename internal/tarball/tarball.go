package tarball

// Based on: https://gist.github.com/indraniel/1a91458984179ab4cf80

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/g2a-com/klio/internal/log"
)

// Extract extracts tar.gz archive into specified directory
func Extract(gzipStream io.Reader, outputDir string) error {
	log.Debugf("Start extracting tarball to %s", outputDir)

	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(uncompressedStream)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		path := filepath.Join(outputDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			log.Spamf("Creating directory: %s", path)
			if err := os.Mkdir(path, 0755); err != nil && !os.IsExist(err) {
				return err
			}
		case tar.TypeReg:
			log.Spamf("Creating file: %s", path)
			outFile, err := os.Create(path)
			if err != nil {
				return err
			}
			defer outFile.Close()
			if runtime.GOOS != "windows" {
				if err = outFile.Chmod(os.FileMode(header.Mode)); err != nil {
					return err
				}
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}

			// Despite previous defer, close this file anyway,
			// it will prevent hitting limit of open files.
			outFile.Close()
		default:
			return fmt.Errorf(
				"tarball contains unknown type: %v in %s",
				header.Typeflag,
				path,
			)
		}
	}

	return nil
}
