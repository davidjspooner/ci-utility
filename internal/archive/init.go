package archive

import (
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

// Commands returns the list of archive-related CLI commands, including checksum and compress.
func Commands() []cmd.Command {

	group := cmd.NewCommandGroup(
		"archive",
		"Archive commands",
	)

	//define the commands for checksum and compress with their respective options.
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
	// Add subcommands to the archive command.
	group.SubCommands().MustAdd(checksumCmd, compressCmd)
	return []cmd.Command{group}
}
