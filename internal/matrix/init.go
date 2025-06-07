package matrix

import (
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

// Commands returns the CLI command group for matrix operations.
func Commands() []cmd.Command {
	// Create the root command group for matrix tools.
	group := cmd.NewCommandGroup(
		"matrix",
		"Tools for an environment matrix ",
	)
	// Define the run command for executing commands over the matrix.
	runCommand := cmd.NewCommand(
		"run",
		"Run a command for each combintation of an environment matrix ",
		doMatrixExecute,
		&RunOptions{},
	)
	// Add the run command to the group.
	group.SubCommands().Add(runCommand)
	// Return the group as a list of commands.
	return []cmd.Command{
		group,
	}
}
