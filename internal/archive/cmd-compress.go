package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type CompressOptions struct {
	Format         string `flag:"--format,Format to compress the files (zip, tar.gz)"`
	RemoveOriginal bool   `flag:"--remove-original,Remove original files after compression"`
}

func compressCommand(ctx context.Context, option *CompressOptions, args []string) error {
	// Check if the correct number of arguments is provided

	var err error
	paths, err := globFiles(args)
	if err != nil {
		return fmt.Errorf("error globbing files: %s", err)
	}
	for _, path := range paths {
		switch option.Format {
		case "zip":
			err = compressToZip(path)
		case "tar.gz":
			err = compressToTarGz(path)
		default:
			// Print an error message if the format is not recognized
			return fmt.Errorf("unsupported --format: %q . Please use 'zip' or 'tar.gz'", option.Format)
		}
		if err != nil {
			return fmt.Errorf("error compressing file %s: %v", path, err)
		}
		slog.DebugContext(ctx, "Compressed file successfully", "path", path, "format", option.Format)
	}
	if option.RemoveOriginal {
		// Call the function to remove original files
		for _, path := range args {
			err = removeOriginal(path)

			if err != nil {
				slog.ErrorContext(ctx, "Error removing original file", "path", path, "error", err)
			}
		}
		slog.DebugContext(ctx, "Original files removed successfully", "paths", paths)
	}
	return nil
}

func compressToZip(path string) error {
	zipFile, err := os.Create(path + ".zip")
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(path, filePath)
		if err != nil {
			return err
		}

		if info.IsDir() {
			if relPath == "." {
				return nil
			}
			_, err := zipWriter.Create(relPath + "/")
			return err
		}

		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

func compressToTarGz(path string) error {
	tarGzFile, err := os.Create(path + ".tar.gz")
	if err != nil {
		return err
	}
	defer tarGzFile.Close()

	gzipWriter := gzip.NewWriter(tarGzFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(path, filePath)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, relPath)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tarWriter, file)
		return err
	})

	return err
}

func removeOriginal(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path %s: %v", path, err)
	}

	if info.IsDir() {
		err = os.RemoveAll(path)
		if err != nil {
			return fmt.Errorf("failed to remove directory %s: %v", path, err)
		}
	} else {
		err = os.Remove(path)
		if err != nil {
			return fmt.Errorf("failed to remove file %s: %v", path, err)
		}
	}

	return nil
}
