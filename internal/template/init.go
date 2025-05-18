package template

import (
	"github.com/davidjspooner/go-text/pkg/cmd"
)

func Commands() []cmd.Command {
	templateCommand := cmd.NewCommand(
		"template",
		"Template commands",
		nil,
		&cmd.NoopOptions{},
	)
	expand := cmd.NewCommand(
		"expand",
		"Expand a template file",
		expandTemplate,
		&expandOptions{},
	)
	cmd2 := cmd.NewCommand(
		"man",
		"Manual for template file syntax",
		templateManPage,
		&cmd.NoopOptions{},
	)
	templateCommand.SubCommands().MustAdd(expand, cmd2)
	return []cmd.Command{templateCommand}
}
