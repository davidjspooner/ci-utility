package review

import "context"

// GoStaticAnalysis runs go vet, staticcheck, etc.
type GoStaticAnalysis struct{}

func (s *GoStaticAnalysis) Name() string { return "go_static_analysis" }
func (s *GoStaticAnalysis) Run(ctx context.Context, meta *ProjectMeta, options *ReviewOptions) ([]Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// TODO: parse output of go vet, staticcheck
	return []Result{{Name: s.Name(), Score: 0, TopIssues: nil}}, nil
}
