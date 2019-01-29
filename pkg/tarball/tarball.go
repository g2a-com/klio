package tarball

// Based on: https://gist.github.com/indraniel/1a91458984179ab4cf80

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"stash.code.g2a.com/CLI/core/pkg/log"
)

// Extract extracts tar.gz archive into specified directory
func Extract(gzipStream io.Reader, outputDir string) error {
	log.Debugf("start extracting tarball to %s", outputDir)

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
			log.Spamf("creating directory: %s", path)
			if err := os.Mkdir(path, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			log.Spamf("creating file: %s", path)
			outFile, err := os.Create(path)
			if err != nil {
				return err
			}
			defer outFile.Close()
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
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
