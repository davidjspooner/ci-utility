package review

import "context"

// GoProjectHygiene ensures project layout and metadata.
type GoProjectHygiene struct{}

func (p *GoProjectHygiene) Name() string { return "go_project_hygiene" }
func (p *GoProjectHygiene) Run(ctx context.Context, meta *ProjectMeta) ([]Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// TODO: check for LICENSE, TODO.md, .gitignore, temp files
	return []Result{{Name: p.Name(), Score: 0, TopIssues: nil}}, nil
}
