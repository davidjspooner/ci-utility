package golang

import (
	"context"
	"slices"
	"strings"
)

// Issue represents a single issue found during a review.
// It contains information about the package, file, line number, and a description of the issue.
// It can also contain nested issues, allowing for hierarchical representation of problems.
type Issue struct {
	Package  string
	Filename string
	Line     int // line number in the file
	Count    int

	Type     string   // e.g., "complexity", "size", "comments", etc.
	Message  string   // description of the issue
	Children []*Issue // nested issues, if any
}

// Result is the output of one Review.
// It contains the name of the review, a score, and a list of top-level issues.
type Result struct {
	Name   string   `yaml:"name"`
	Score  int      `yaml:"score"`
	Issues []*Issue `yaml:"top_issues"`
}

// Summerize aggregates issues in the Result.
// It groups issues by filename and type, counting occurrences and nesting them as children.
func (r *Result) Summerize() {
	// Set the score to the number of issues before summarization.
	r.Score = len(r.Issues)
	parents := map[string]*Issue{}
	for _, issue := range r.Issues {
		// Group issues by filename and type.
		key := issue.Filename + "|" + issue.Type
		parent, ok := parents[key]
		if !ok {
			parent = &Issue{
				Filename: issue.Filename,
				Type:     issue.Type,
			}
			parents[key] = parent
		}
		// Add the issue as a child and increment the count.
		parent.Children = append(parent.Children, issue)
		parent.Count += issue.Count
	}
	// Rebuild the Issues slice with summarized parents.
	r.Issues = make([]*Issue, 0, len(parents))
	for _, parent := range parents {
		if len(parent.Children) == 1 {
			r.Issues = append(r.Issues, parent.Children[0])
			continue
		}
		r.Issues = append(r.Issues, parent)
	}
	// Sort issues by count, filename, type, and line.
	slices.SortFunc(r.Issues, func(a, b *Issue) int {
		r := a.Count - b.Count
		if r == 0 {
			r = strings.Compare(a.Filename, b.Filename)
		}
		if r == 0 {
			r = strings.Compare(a.Type, b.Type)
		}
		if r == 0 {
			r = a.Line - b.Line
		}
		return -r
	})
	// Update the score after summarization.
	r.Score = len(r.Issues)
}

// Category is the interface that each Category module must implement.
// It defines a Name method to return the category name and a Run method to execute the review logic.
type Category interface {
	Name() string
	Run(ctx context.Context, meta *Meta, options *ReviewOptions) ([]*Result, error)
}

// Meta contains data passed to Reviews.
// It includes root paths and can be extended to include more metadata like AST, coverage, etc.
type Meta struct {
	RootPaths []string
}
