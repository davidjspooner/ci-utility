package github

import (
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

func Commands() []cmd.Command {
	githubCommand := cmd.NewCommand(
		"github",
		"GitHub commands",
		nil,
		&cmd.NoopOptions{},
	)
	cmd1 := cmd.NewCommand(
		"release",
		"Create a GitHub release",
		executeGithubRelease,
		&GithubReleaseOptions{},
	)
	cmd2 := cmd.NewCommand(
		"update-pull-request",
		"Update a GitHub pull request with the latest changes from the base branch",
		executeUpdateGithubPRMeta,
		&GithubPRUpdateOptions{},
	)
	githubCommand.SubCommands().MustAdd(cmd1, cmd2)
	return []cmd.Command{githubCommand}
}
