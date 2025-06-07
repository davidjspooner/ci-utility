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

// CompressOptions holds options for the compress command.
type CompressOptions struct {
	Format         string `flag:"--format,Format to compress the files (zip, tar.gz)"`
	Target         string `flag:"--target,Combine multiple files into a single archive"`
	Rename         string `flag:"--rename,Rename the file inside the archive (cannot use with --target)"`
	RemoveOriginal bool   `flag:"--remove-original,Remove original files after compression"`
}

// compressCommand compresses files or directories according to the provided options.
// It supports both zip and tar.gz formats, and can optionally remove the original files.
func compressCommand(ctx context.Context, options *CompressOptions, args []string) error {
	// Check if the correct number of arguments is provided

	var err error
	// Expand globs in the input arguments to get the actual file paths.
	paths, err := globFiles(args)
	if err != nil {
		return fmt.Errorf("error globbing files: %s", err)
	}
	if len(paths) == 0 {
		return fmt.Errorf("no files or directories found to compress")
	}

	if options.Target != "" {
		if options.Rename != "" {
			return fmt.Errorf("cannot use --rename with multiple files, please specify a single file")
		}

		// Ensure the target directory exists.
		dir := filepath.Dir(options.Target)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("error creating target directory: %s", err)
		}
	}

	// Select compression format based on user option.
	switch options.Format {
	case "zip":
		if options.Target == "" {
			for _, path := range paths {
				// If no target is specified, compress each file individually.
				options.Target = path + ".zip"
				err = compressToZip(ctx, options, []string{path})
				if err != nil {
					return fmt.Errorf("error compressing file %s: %v", path, err)
				}
			}
		} else {
			err = compressToZip(ctx, options, paths)
			if err != nil {
				return fmt.Errorf("error compressing files to %s: %v", options.Target, err)
			}
		}
	case "tar.gz":
		if options.Target == "" {
			for _, path := range paths {
				// If no target is specified, compress each file individually.
				options.Target = path + ".zip"
				err = compressToTarGz(ctx, options, []string{path})
				if err != nil {
					return fmt.Errorf("error compressing file %s: %v", path, err)
				}
			}
		} else {
			err = compressToTarGz(ctx, options, paths)
			if err != nil {
				return fmt.Errorf("error compressing files to %s: %v", options.Target, err)
			}
		}
	default:
		// Print an error message if the format is not recognized
		return fmt.Errorf("unsupported --format: %q . Please use 'zip' or 'tar.gz'", options.Format)
	}
	if err != nil {
		return fmt.Errorf("error compressing files: %v", err)
	}
	// Remove original files if requested.
	if options.RemoveOriginal {
		// Call the function to remove original files
		for _, path := range args {
			err = removeOriginal(path)

			if err != nil {
				slog.ErrorContext(ctx, "Error removing original file", "path", path, "error", err)
			}
		}
		// Log successful removal of original files.
		slog.InfoContext(ctx, "Original files removed successfully", "paths", paths)
	}
	return nil
}

// compressToZip compresses the given path into a .zip archive.
// It walks the directory tree and adds all files and directories to the archive.
func compressToZip(ctx context.Context, options *CompressOptions, paths []string) error {
	outFile, err := os.Create(options.Target)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %v", err)
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	for _, root := range paths {
		err = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			relPath, err := filepath.Rel(filepath.Dir(root), path)
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
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			fh, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			fh.Name = relPath
			if options.Rename != "" && len(paths) == 1 {
				fh.Name = options.Rename
			}
			fh.Method = zip.Deflate

			writer, err := zipWriter.CreateHeader(fh)
			if err != nil {
				return err
			}
			_, err = io.Copy(writer, file)
			return err
		})
		if err != nil {
			return err
		}
	}
	slog.Info("Created zip archive", "target", options.Target)
	return nil
}

// compressToTarGz compresses the given path into a .tar.gz archive.
// It walks the directory tree and adds all files and directories to the archive.
func compressToTarGz(ctx context.Context, options *CompressOptions, paths []string) error {
	outFile, err := os.Create(options.Target)
	if err != nil {
		return fmt.Errorf("failed to create tar.gz file: %v", err)
	}
	defer outFile.Close()

	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	for _, root := range paths {
		err = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			relPath, err := filepath.Rel(filepath.Dir(root), path)
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			header.Name = relPath
			if options.Rename != "" && len(paths) == 1 {
				header.Name = options.Rename
			}
			if err := tarWriter.WriteHeader(header); err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarWriter, file)
			return err
		})
		if err != nil {
			return err
		}
	}
	slog.Info("Created tar.gz archive", "target", options.Target)
	return nil
}

// removeOriginal deletes the original file or directory at the given path.
// It removes directories recursively and files directly.
func removeOriginal(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path %s: %v", path, err)
	}

	// Remove directory or file as appropriate.
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
