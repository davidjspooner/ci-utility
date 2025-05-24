package git

import (
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

func Commands() []cmd.Command {

	gitCommand := cmd.NewCommand(
		"git",
		"Git commands",
		nil,
		&cmd.NoopOptions{},
	)

	cmd1 := cmd.NewCommand(
		"suggest-build-env",
		"Get the environment variables for the current build",
		executeGetGitEnv,
		&GetGitEnvOptions{},
	)
	cmd2 := cmd.NewCommand(
		"update-tag",
		"Automatically increment Git tags based on commit messages (e.g., fix:, feat:, breaking:)",
		executeBumpGitTag,
		&BumpGitTagOptions{
			Remote: "origin",
			Prefix: "v",
		},
	)

	gitCommand.SubCommands().MustAdd(cmd1, cmd2)
	return []cmd.Command{gitCommand}
}
