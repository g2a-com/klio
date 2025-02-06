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
	"github.com/spf13/afero"
)

// Extract extracts tar.gz archive into specified directory.
func Extract(gzipStream io.Reader, fs afero.Fs, outputDir string) error {
	log.Debugf("Start extracting tarball to %s", outputDir)

	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
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
			if err := fs.Mkdir(path, 0o755); err != nil && !os.IsExist(err) {
				return err
			}
		case tar.TypeReg:
			log.Spamf("Creating file: %s", path)
			outFile, err := fs.Create(path)
			if err != nil {
				return err
			}
			defer func() { _ = outFile.Close() }()
			if runtime.GOOS != "windows" {
				if err = fs.Chmod(path, os.FileMode(header.Mode)); err != nil {
					return err
				}
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}

			// Despite previous defer, close this file anyway,
			// it will prevent hitting limit of open files.
			_ = outFile.Close()
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
