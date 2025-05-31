package review

import "context"

// Result is the output of one Review.
type Result struct {
	Name      string   `yaml:"name"`
	Score     int      `yaml:"score"`
	TopIssues []string `yaml:"top_issues"`
}

// Category is the interface that each Category module must implement.
type Category interface {
	Name() string
	Run(ctx context.Context, meta *ProjectMeta, options *ReviewOptions) ([]Result, error)
}

// ProjectMeta contains data passed to Reviews.
type ProjectMeta struct {
	RootPath string
	// TODO: include more parsed metadata, AST, coverage, etc.
}
