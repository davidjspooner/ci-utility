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
	releaseCreate := cmd.NewCommand(
		"create",
		"Create a GitHub release",
		executeGithubReleaseCreate,
		&GithubReleaseCreateOptions{},
	)
	prUpdate := cmd.NewCommand(
		"update",
		"Update a GitHub pull request with the latest changes from the base branch",
		executeUpdateGithubPRMeta,
		&GithubPRUpdateOptions{},
	)

	pullRequest := cmd.NewCommand(
		"pull-request",
		"GitHub pull request commands",
		nil,
		&cmd.NoopOptions{},
	)
	release := cmd.NewCommand(
		"release",
		"GitHub release commands",
		nil,
		&cmd.NoopOptions{},
	)

	pullRequest.SubCommands().MustAdd(prUpdate)
	release.SubCommands().MustAdd(releaseCreate)

	githubCommand.SubCommands().MustAdd(release, pullRequest)
	return []cmd.Command{githubCommand}
}
