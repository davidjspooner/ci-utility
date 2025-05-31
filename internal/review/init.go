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
	RootPath    string `flag:"--root-path,Root directory of the project to review"`
	ReportPath  string `flag:"--report-path,Path to save the review report"`
	TargetScore int    `flag:"--target-score,Target score for the review"`
	Tolerate    string `flag:"--tolerate,Comma seperated list of categories to tolerate missing the target score"`
	Skip        string `flag:"--skip,Comma seperated list of categories to skip in the review"`
}

func Commands() []cmd.Command {
	reviewCommand := cmd.NewCommand(
		"review",
		"review proejct and generare a review.yaml file",
		func(ctx context.Context, options *ReviewOptions, args []string) error {
			meta := ProjectMeta{
				RootPath: options.RootPath,
			}
			results, err := registry.Run(ctx, &meta, options)
			if err != nil {
				return err
			}

			slices.SortFunc(results, func(a, b Result) int {
				return a.Score - b.Score
			})

			var failures []string
			for _, result := range results {
				if result.Score < options.TargetScore {
					failures = append(failures, fmt.Sprintf("%s: %d", result.Name, result.Score))
				}
			}

			f, err := os.Create(options.ReportPath)
			if err != nil {
				return err
			}
			defer f.Close()
			e := yaml.NewEncoder(f)
			defer e.Close()
			if err := e.Encode(results); err != nil {
				return err
			}

			fmt.Printf("Generated %s \n", options.ReportPath)
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
			RootPath:   ".",
			ReportPath: "review.yaml",
		},
	)
	return []cmd.Command{
		reviewCommand,
	}
}
