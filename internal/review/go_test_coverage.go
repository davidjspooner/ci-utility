package review

import "context"

// GoTestCoverage checks code coverage.
type GoTestCoverage struct{}

func (t *GoTestCoverage) Name() string { return "go_test_coverage" }
func (t *GoTestCoverage) Run(ctx context.Context, meta *ProjectMeta) ([]Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// TODO: implement test coverage scan
	return []Result{{Name: t.Name(), Score: 0, TopIssues: nil}}, nil
}
