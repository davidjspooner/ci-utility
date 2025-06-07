package github

import (
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

// Commands returns the list of GitHub-related CLI commands for the application.
func Commands() []cmd.Command {
	githubCommand := cmd.NewCommandGroup(
		"github",
		"GitHub commands",
	)
	// Create the release create command.
	releaseCreate := cmd.NewCommand(
		"create",
		"Create a GitHub release",
		executeGithubReleaseCreate,
		&ReleaseCreateOptions{},
	)
	// Create the PR update command.
	prUpdate := cmd.NewCommand(
		"update",
		"Update a GitHub pull request with the latest changes from the base branch",
		executeUpdateGithubPRMeta,
		&PRUpdateOptions{},
	)

	// Create command groups for pull requests and releases.
	pullRequest := cmd.NewCommandGroup(
		"pull-request",
		"GitHub pull request commands",
	)
	release := cmd.NewCommandGroup(
		"release",
		"GitHub release commands",
	)

	// Add subcommands to their respective groups.
	pullRequest.SubCommands().MustAdd(prUpdate)
	release.SubCommands().MustAdd(releaseCreate)

	// Add groups to the root github command.
	githubCommand.SubCommands().MustAdd(release, pullRequest)
	return []cmd.Command{githubCommand}
}
