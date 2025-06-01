package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/davidjspooner/ci-utility/internal/archive"
	"github.com/davidjspooner/ci-utility/internal/git"
	"github.com/davidjspooner/ci-utility/internal/github"
	"github.com/davidjspooner/ci-utility/internal/matrix"
	"github.com/davidjspooner/ci-utility/internal/review"
	"github.com/davidjspooner/ci-utility/internal/template"
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

type GlobalOptions struct {
	cmd.LogOptions
}

func main() {
	root := cmd.NewCommand("", "A utility for CI/CD operations",
		func(ctx context.Context, options *GlobalOptions, args []string) error {
			err := options.LogOptions.SetupSLOG()
			if err != nil {
				slog.ErrorContext(ctx, "failed to setup logger", "error", err)
				return err
			}
			err = cmd.ShowHelpForMissingSubcommand(ctx)
			return err
		}, &GlobalOptions{LogOptions: cmd.LogOptions{Level: "info"}},
		cmd.LogicalGroup)

	cmd.RootCommand = root
	versionCommand := cmd.VersionCommand()
	gitCommands := git.Commands()
	archiveCommands := archive.Commands()
	githubCommands := github.Commands()
	templateCommands := template.Commands()
	reviewCommands := review.Commands()
	matrixCommands := matrix.Commands()

	subcommands := cmd.RootCommand.SubCommands()
	subcommands.MustAdd(
		versionCommand,
		gitCommands,
		archiveCommands,
		githubCommands,
		templateCommands,
		reviewCommands,
		matrixCommands,
	)

	ctx := context.Background()
	err := cmd.Run(ctx, os.Args[1:])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}
}
