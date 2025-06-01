package matrix

import (
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

func Commands() []cmd.Command {
	group := cmd.NewGroupCommand(
		"matrix",
		"Tools for an environment matrix ",
	)
	runCommand := cmd.NewCommand(
		"run",
		"Run a command for each combintation of an environment matrix ",
		doMatrixExecute,
		&MatrixRunOptions{},
	)
	group.SubCommands().Add(runCommand)
	return []cmd.Command{
		group,
	}
}
