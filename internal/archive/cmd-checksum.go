package archive

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
)

// ChecksumOptions holds options for the checksum command.
type ChecksumOptions struct {
	Algorithm    string `flag:"--algorithm,Checksum algorithm (e.g., sha256, md5)"`
	Extension    string `flag:"--extension,File extension for individual checksum files"`
	CombinedFile string `flag:"--combined-file,Write all checksums to a single file"`
}

// executeChecksum generates checksums for the specified files using the provided options.
// It supports writing checksums to individual files or a combined file.
func executeChecksum(ctx context.Context, option *ChecksumOptions, args []string) error {
	// Validate the inputs
	if len(args) < 1 {
		return fmt.Errorf("no files specified")
	}
	files, err := globFiles(args)
	if err != nil {
		return fmt.Errorf("error globbing files: %s", err)
	}
	if option.Extension == "" && option.CombinedFile == "" {
		return fmt.Errorf("need to specify --extension and/or --combined-file")
	}

	// if the combined file is specified, create it.
	var combinedFile *os.File
	if option.CombinedFile != "" {
		combinedFile, err = os.Create(option.CombinedFile)
		if err != nil {
			return fmt.Errorf("failed to create combined checksum file: %v", err)
		}
		defer combinedFile.Close()
	}

	// Generate checksums for each file.
	for _, file := range files {
		checksum, err := generateChecksum(file, option.Algorithm)
		if err != nil {
			return fmt.Errorf("error generating checksum for %s: %v", file, err)
		}
		baseFile := path.Base(file)
		dirName := path.Dir(file)
		// If the extension is specified, create a separate checksum file.
		if option.Extension != "" {
			seperateFile, err := os.Create(path.Join(dirName, baseFile) + option.Extension)
			if err != nil {
				return fmt.Errorf("failed to create checksum file: %v", err)
			}
			_, err = fmt.Fprintf(seperateFile, "%s  %s\n", checksum, baseFile)
			slog.DebugContext(ctx, "Checksum", "checksum", checksum, "file", baseFile)
			seperateFile.Close()
			if err != nil {
				return fmt.Errorf("failed to write checksum file: %v", err)
			}
		}
		// additionally write to the combined file if specified.
		if option.CombinedFile != "" {
			_, err = fmt.Fprintf(combinedFile, "%s  %s\n", checksum, baseFile)
			if err != nil {
				return fmt.Errorf("failed to write combined checksum file: %v", err)
			}
		}
	}

	return nil
}

// generateChecksum computes the checksum of a file using the specified algorithm.
// It returns the checksum as a hex string, or an error if the file is invalid or the algorithm is unsupported.
func generateChecksum(file string, algorithm string) (string, error) {
	stat, err := os.Stat(file)
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %v", err)
	}
	if stat.IsDir() {
		return "", fmt.Errorf("file is a directory: %s", file)
	}
	if stat.Size() == 0 {
		return "", fmt.Errorf("file is empty: %s", file)
	}
	input, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	var checksum string
	defer input.Close()
	switch algorithm {
	case "sha256":
		// Generate sha256 checksum.
		h := sha256.New()
		if _, err := io.Copy(h, input); err != nil {
			return "", fmt.Errorf("failed to generate checksum: %v", err)
		}
		checksum = fmt.Sprintf("%x", h.Sum(nil))
	case "md5":
		// Generate md5 checksum.
		h := md5.New()
		if _, err := io.Copy(h, input); err != nil {
			return "", fmt.Errorf("failed to generate checksum: %v", err)
		}
		checksum = fmt.Sprintf("%x", h.Sum(nil))
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
	if checksum == "" {
		return "", fmt.Errorf("failed to generate checksum")
	}
	return checksum, err
}
