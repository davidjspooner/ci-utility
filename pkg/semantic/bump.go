package semantic

import "strings"

type Bump struct {
	Level string
	Hints []string
}

type BumpArray []Bump

// Bump levels
var Bumps = BumpArray{
	{Level: "major", Hints: []string{"BREAKING CHANGE", "breaking:"}},
	{Level: "minor", Hints: []string{"feat:"}},
	{Level: "patch", Hints: []string{"fix:", "chore:", "docs:", "style:", "refactor:", "perf:", "test:"}},
}

func (bumps BumpArray) GetVersionBump(commits []string) (string, error) {
	found := map[string]bool{}
	for _, msg := range commits {
		for _, bump := range bumps {
			for _, hint := range bump.Hints {
				if strings.Contains(msg, hint) {
					found[bump.Level] = true
				}
			}
		}
	}

	for _, bump := range bumps {
		if found[bump.Level] {
			return bump.Level, nil
		}
	}

	return "patch", nil
}
