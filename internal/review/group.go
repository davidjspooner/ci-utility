package review

import (
	"context"
	"fmt"
)

// Group holds and runs Reviews.
type Group struct {
	Reviews []Category
}

// Register adds a new Review to the registry.
func (r *Group) Register(a Category) {
	r.Reviews = append(r.Reviews, a)
}

// FindByName returns a Review by its name.
func (r *Group) FindByName(name string) Category {
	for _, a := range r.Reviews {
		if a.Name() == name {
			return a
		}
	}
	return nil
}

// RunAll executes all registered Reviews.
func (r *Group) Run(ctx context.Context, meta *ProjectMeta, options *ReviewOptions) ([]Result, error) {
	var results []Result
	for _, a := range r.Reviews {
		results, err := a.Run(ctx, meta, options)
		if err != nil {
			continue
		}
		for _, result := range results {
			if r.FindByName(result.Name) != nil {
				return nil, fmt.Errorf("duplicate review name found: %s", result.Name)
			}
			results = append(results, result)
		}
	}
	return results, nil
}

var registry = &Group{}

// RegisterDefaults preloads the registry with all built-in checks.
func init() {
	registry.Register(&GeneralProjectHygiene{})
	registry.Register(&GoErrorHandling{})
	registry.Register(&GoDocumentation{})
	registry.Register(&GoTestCoverage{})
	registry.Register(&GoCodeStructure{})
	registry.Register(&GoStaticAnalysis{})
}
