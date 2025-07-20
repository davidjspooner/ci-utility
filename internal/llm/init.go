package llm

import (
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

type LLMOptions struct {
	// Add any options specific to the review command here
	Config     string `flag:"--config,config file for LLM model to use"`
	OutputFile string `flag:"--output,output file for LLM response"`
	ToolsFile  string `flag:"--tools,used to defined tools the LLM can use"`
}

// Commands returns the list of git-related CLI commands for the application.
func AddCommandsTo(parent cmd.Command) error {
	llm := cmd.NewCommand(
		"llm",
		"LLM utility commands",
		nil, // No specific handler for the llm command itself
		&LLMOptions{
			Config:     ".llm-config.yaml",
			OutputFile: "-", // Default to stdout
		},
	)

	customCmd := cmd.NewCommand(
		"run|execute",
		"Run LLM with specified options",
		executeCustomCommand,
		&CustomOptions{},
	)

	toolsCmd := cmd.NewCommandGroup(
		"tools",
		"Manage tools definition for use by custom LLM commands",
	)

	addToolsCmd := cmd.NewCommand(
		"add|define|update",
		"Add or update tools definition for use by custom LLM commands",
		addToolsDefCommand,
		&ModifyToolsDefOptions{},
	)
	removeToolsDefCmd := cmd.NewCommand(
		"remove|delete",
		"Remove tools definition for use by custom LLM commands",
		removeToolsDefCommand,
		&ModifyToolsDefOptions{},
	)

	toolsCmd.SubCommands().Add(addToolsCmd, removeToolsDefCmd)
	llm.SubCommands().Add(customCmd) //, toolsCmd is work in progress
	parent.SubCommands().Add(llm)

	return nil
}
