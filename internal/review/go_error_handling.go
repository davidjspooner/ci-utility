package review

import "context"

// GoErrorHandling evaluates error practices.
type GoErrorHandling struct{}

func (e *GoErrorHandling) Name() string { return "go_error_handling" }
func (e *GoErrorHandling) Run(ctx context.Context, meta *ProjectMeta, options *ReviewOptions) ([]Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// TODO: check for proper error wrapping, MustX, Unwrap()
	return []Result{{Name: e.Name(), Score: 0, TopIssues: nil}}, nil
}
