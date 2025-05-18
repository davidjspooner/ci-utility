package archive

import (
	"github.com/davidjspooner/go-text/pkg/cmd"
)

func Commands() []cmd.Command {

	archiveCommand := cmd.NewCommand(
		"archive",
		"Archive commands",
		nil,
		&cmd.NoopOptions{},
	)

	checksumCmd := cmd.NewCommand(
		"checksum",
		"Generate checksum(s) for file(s) using a specified algorithm",
		executeChecksum,
		&ChecksumOptions{
			Algorithm: "sha256",
		},
	)
	compressCmd := cmd.NewCommand(
		"compress",
		"Compress files or directories into zip or tar.gz formats",
		compressCommand,
		&CompressOptions{
			Format:         "tar.gz",
			RemoveOriginal: false,
		},
	)
	archiveCommand.SubCommands().MustAdd(checksumCmd, compressCmd)
	return []cmd.Command{archiveCommand}
}
