package llm

import (
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

// Commands returns the list of git-related CLI commands for the application.
func AddCommandsTo(parent cmd.Command) error {
	llm := cmd.NewCommandGroup(
		"llm",
		"LLM utility commands",
	)

	customCmd := cmd.NewCommand(
		"run|execute",
		"Run LLM with specified options",
		executeCustomCommand,
		&CustomOptions{
			Config:     ".llm-config.yaml",
			OutputFile: "-", // Default to stdout
		},
	)

	toolsCmd := cmd.NewCommandGroup(
		"tools",
		"Manage tools definition for use by custom LLM commands",
	)

	llm.SubCommands().Add(customCmd, toolsCmd)

	addToolsCmd := cmd.NewCommand(
		"add|define|update",
		"Add or update tools definition for use by custom LLM commands",
		addToolsDefCommand,
		&ModifyToolsDefOptions{
			Config:     ".llm-config.yaml",
			OutputFile: "-", // Default to stdout
		},
	)
	removeToolsDefCmd := cmd.NewCommand(
		"remove|delete",
		"Remove tools definition for use by custom LLM commands",
		removeToolsDefCommand,
		&ModifyToolsDefOptions{
			Config:     ".llm-config.yaml",
			OutputFile: "-", // Default to stdout
		},
	)

	toolsCmd.SubCommands().Add(addToolsCmd, removeToolsDefCmd)
	llm.SubCommands().Add(customCmd, toolsCmd)

	return nil
}
