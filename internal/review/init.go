package review

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/davidjspooner/go-text-cli/pkg/cmd"
	"gopkg.in/yaml.v3"
)

type ReviewOptions struct {
	// Add any options specific to the review command here
	RootDir     string `flag:"--root-dir,Root directory of the project to review"`
	Report      string `flag:"--report,Path to save the review report"`
	TargetScore int    `flag:"--target-score,Target score for the review"`
}

func Commands() []cmd.Command {
	reviewCommand := cmd.NewCommand(
		"review",
		"Review proejct and optionally generate a review.yaml file",
		func(ctx context.Context, options *ReviewOptions, args []string) error {
			meta := ProjectMeta{
				RootPath: options.RootDir,
			}
			results, err := registry.Run(ctx, &meta, options)
			if err != nil {
				return err
			}

			slices.SortFunc(results, func(a, b *Result) int {
				return a.Score - b.Score
			})

			var failures []string
			for _, result := range results {
				if result.Score < options.TargetScore {
					failures = append(failures, fmt.Sprintf("%s: %d", result.Name, result.Score))
				}
			}

			if options.Report != "" {
				f, err := os.Create(options.Report)
				if err != nil {
					return err
				}
				defer f.Close()
				e := yaml.NewEncoder(f)
				defer e.Close()
				if err := e.Encode(results); err != nil {
					return err
				}

				fmt.Printf("Generated %s \n", options.Report)
			}

			if len(failures) > 0 {
				fmt.Printf("The following missed the target of %d:\n", options.TargetScore)
				for _, failure := range failures {
					fmt.Printf("- %s\n", failure)
				}
				return fmt.Errorf("review failed with %d issues", len(failures))
			}

			return nil
		},
		&ReviewOptions{
			RootDir: ".",
		},
	)
	return []cmd.Command{
		reviewCommand,
	}
}
