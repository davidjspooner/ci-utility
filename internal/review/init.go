package review

import (
	"context"

	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

type ReviewOptions struct {
	// Add any options specific to the review command here
}

func Commands() []cmd.Command {
	reviewCommand := cmd.NewCommand(
		"review",
		"review proejct and generare a review.yaml file",
		func(ctx context.Context, options *ReviewOptions, args []string) error {
			// TODO Implementation of the review command goes here
			return nil
		},
		&ReviewOptions{},
	)
	return []cmd.Command{
		reviewCommand,
	}
}
