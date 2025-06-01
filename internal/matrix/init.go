package matrix

import (
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

func Commands() []cmd.Command {
	group := cmd.NewGroupCommand(
		"matrix",
		"Run commands for an environment matrix ",
	)
	runCommand := cmd.NewCommand(
		"run",
		"Run commands for an environment matrix ",
		executeMatrixRun,
		&MatrixRunOptions{},
	)
	group.SubCommands().Add(runCommand)
	return []cmd.Command{
		group,
	}
}
