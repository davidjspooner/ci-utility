package review

import "context"

// GoBuildReadiness verifies cross-platform build and tagging.
type GoBuildReadiness struct{}

func (b *GoBuildReadiness) Name() string { return "go_build_readiness" }
func (b *GoBuildReadiness) Run(ctx context.Context, meta *ProjectMeta) ([]Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// TODO: test go build/install on all GOOS/GOARCH
	return []Result{{Name: b.Name(), Score: 0, TopIssues: nil}}, nil
}
