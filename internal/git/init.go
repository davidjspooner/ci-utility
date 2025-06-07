package git

import (
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

// Commands returns the list of git-related CLI commands for the application.
func Commands() []cmd.Command {

	// Create the root git command.
	gitCommand := cmd.NewCommand(
		"git",
		"Git commands",
		nil,
		&cmd.NoopOptions{},
	)

	// Define the subcommand for suggesting build environment variables.
	cmd1 := cmd.NewCommand(
		"suggest-build-env",
		"Get the environment variables for the current build",
		executeSuggestBuildEnv,
		&SuggestBuildEnvOptions{},
	)
	// Define the subcommand for updating git tags automatically.
	cmd2 := cmd.NewCommand(
		"update-tag",
		"Automatically increment Git tags based on commit messages (e.g., fix:, feat:, breaking:)",
		executeBumpGitTag,
		&BumpGitTagOptions{
			Remote: "origin",
			Prefix: "v",
		},
	)

	// Add subcommands to the root git command.
	gitCommand.SubCommands().MustAdd(cmd1, cmd2)
	return []cmd.Command{gitCommand}
}
