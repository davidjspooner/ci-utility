package golang

import (
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

// Commands returns the list of Go-related CLI commands for the application.
func Commands() []cmd.Command {
	group := cmd.NewCommandGroup("go", "go related commands")
	// Add the review command as a subcommand.
	group.SubCommands().Add(reviewCommand)
	return []cmd.Command{
		group,
	}
}
