package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/davidjspooner/ci-utility/internal/archive"
	"github.com/davidjspooner/ci-utility/internal/git"
	"github.com/davidjspooner/ci-utility/internal/github"
	"github.com/davidjspooner/ci-utility/internal/golang"
	"github.com/davidjspooner/ci-utility/internal/llm"
	"github.com/davidjspooner/ci-utility/internal/matrix"
	"github.com/davidjspooner/ci-utility/internal/template"
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

// GlobalOptions holds global CLI options, including logging configuration.
type GlobalOptions struct {
	cmd.LogOptions
}

// main is the entry point for the ci-utility CLI application.
// It sets up the root command, subcommands, and runs the CLI.
func main() {
	root := cmd.NewCommand("", "A utility for CI/CD operations",
		func(ctx context.Context, options *GlobalOptions, args []string) error {
			// Setup the logger based on global options.
			err := options.LogOptions.SetupSLOG()
			if err != nil {
				slog.ErrorContext(ctx, "failed to setup logger", "error", err)
				return err
			}
			// Show help if no subcommand is provided.
			err = cmd.ShowHelpForMissingSubcommand(ctx)
			return err
		}, &GlobalOptions{})

	cmd.Root = root

	// Create sub commands for the root command

	git.AddCommandsTo(cmd.Root)
	archive.AddCommandsTo(cmd.Root)
	github.AddCommandsTo(cmd.Root)
	golang.AddCommandsTo(cmd.Root)
	template.AddCommandsTo(cmd.Root)
	matrix.AddCommandsTo(cmd.Root)
	llm.AddCommandsTo(cmd.Root)

	cmd.Root.SubCommands().Add(cmd.VersionCommand())

	ctx := context.Background()
	// Run the CLI with the provided arguments.
	err := cmd.Run(ctx, os.Args[1:])
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		os.Exit(1)
	}
}
