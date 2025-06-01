package review

import (
	"context"
	"fmt"
)

// Group holds and runs Reviews.
type Group struct {
	Category []Category
}

// Register adds a new Review to the registry.
func (r *Group) Register(a Category) {
	r.Category = append(r.Category, a)
}

// FindByName returns a Review by its name.
func (r *Group) FindByName(name string) Category {
	for _, a := range r.Category {
		if a.Name() == name {
			return a
		}
	}
	return nil
}

// RunAll executes all registered Reviews.
func (r *Group) Run(ctx context.Context, meta *ProjectMeta, options *ReviewOptions) ([]*Result, error) {
	var allResults []*Result
	for _, a := range r.Category {
		results, err := a.Run(ctx, meta, options)
		if err != nil {
			continue
		}
		for _, result := range results {
			for _, seen := range allResults {
				if seen.Name == result.Name {
					// If the result is already seen, skip it
					return nil, fmt.Errorf("duplicate result found: %s", result.Name)
				}
			}
			allResults = append(allResults, result)
		}
	}
	return allResults, nil
}

var registry = &Group{}

// RegisterDefaults preloads the registry with all built-in checks.
func init() {

	registry.Register(&GeneralProjectHygiene{})
	registry.Register(&GoRoot{})
}
