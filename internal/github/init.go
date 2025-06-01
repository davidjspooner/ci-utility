package github

import (
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

func Commands() []cmd.Command {
	githubCommand := cmd.NewGroupCommand(
		"github",
		"GitHub commands",
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

	pullRequest := cmd.NewGroupCommand(
		"pull-request",
		"GitHub pull request commands",
	)
	release := cmd.NewGroupCommand(
		"release",
		"GitHub release commands",
	)

	pullRequest.SubCommands().MustAdd(prUpdate)
	release.SubCommands().MustAdd(releaseCreate)

	githubCommand.SubCommands().MustAdd(release, pullRequest)
	return []cmd.Command{githubCommand}
}
