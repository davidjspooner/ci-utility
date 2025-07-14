package golang

import (
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

// Commands returns the list of Go-related CLI commands for the application.
func AddCommandsTo(parent cmd.Command) error {
	group := cmd.NewCommandGroup("go", "go related commands")
	// Add the review command as a subcommand.
	group.SubCommands().Add(reviewCommand)
	parent.SubCommands().MustAdd(group)
	return nil
}
