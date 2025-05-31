package review

import "context"

// GoDocumentation checks documentation completeness.
type GoDocumentation struct{}

func (d *GoDocumentation) Name() string { return "go_documentation" }
func (d *GoDocumentation) Run(ctx context.Context, meta *ProjectMeta) ([]Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// TODO: implement GoDoc and README analysis
	return []Result{{Name: d.Name(), Score: 0, TopIssues: nil}}, nil
}
