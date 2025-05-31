package review

import "context"

// GeneralProjectHygiene ensures project layout and metadata.
type GeneralProjectHygiene struct{}

func (p *GeneralProjectHygiene) Name() string { return "go_project_hygiene" }
func (p *GeneralProjectHygiene) Run(ctx context.Context, meta *ProjectMeta, options *ReviewOptions) ([]Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// TODO: check for LICENSE, .gitignore, temp files, secrets
	return []Result{{Name: p.Name(), Score: 0, TopIssues: nil}}, nil
}
