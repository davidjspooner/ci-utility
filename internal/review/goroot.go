package review

import (
	"context"
	"fmt"
)

// GoRoot analyzes function size and complexity.
type GoRoot struct{}

var _ Category = (*GoRoot)(nil)

func (c *GoRoot) Name() string { return "go_code_structure" }
func (c *GoRoot) Run(ctx context.Context, meta *ProjectMeta, options *ReviewOptions) ([]*Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	pi, err := scanPackages(meta.RootPath)
	if err != nil {
		return nil, err
	}
	for _, p := range pi.Packages {
		fmt.Printf("Package: %s (%s)\n", p.Name, p.DirPath)
	}

	// TODO: implement cyclomatic complexity, package cohesion checks
	return []*Result{{Name: c.Name(), Score: 0, TopIssues: nil}}, nil
}
