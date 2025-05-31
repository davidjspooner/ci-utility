package review

import "context"

// GoCodeStructure analyzes function size and complexity.
type GoCodeStructure struct{}

func (c *GoCodeStructure) Name() string { return "go_code_structure" }
func (c *GoCodeStructure) Run(ctx context.Context, meta *ProjectMeta) ([]Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// TODO: implement cyclomatic complexity, package cohesion checks
	return []Result{{Name: c.Name(), Score: 0, TopIssues: nil}}, nil
}
