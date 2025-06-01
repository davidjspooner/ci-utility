package review

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
)

type goPackageInfo struct {
	DirPath string
	Name    string
}

type goPackageInventory struct {
	CLIs     []goPackageInfo
	Packages []goPackageInfo
	Tests    []goPackageInfo
}

// scanPackages walks a project root and builds categorized package inventory.
func scanPackages(root string) (*goPackageInventory, error) {
	inventory := &goPackageInventory{}
	visited := make(map[string]map[string]bool) // dir -> pkg names

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		fset := token.NewFileSet()
		f, parseErr := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if parseErr != nil || f == nil {
			return nil
		}

		dir := filepath.Dir(path)
		pkgName := f.Name.Name

		info := goPackageInfo{DirPath: dir, Name: pkgName}

		if strings.HasSuffix(path, "_test.go") {
			if visited[dir] == nil {
				visited[dir] = make(map[string]bool)
			}
			if !visited[dir][pkgName] {
				visited[dir][pkgName] = true
				inventory.Tests = append(inventory.Tests, info)
			}
			return nil
		}

		if visited[dir] == nil {
			visited[dir] = make(map[string]bool)
		}
		if visited[dir][pkgName] {
			return nil // already classified
		}
		visited[dir][pkgName] = true

		if pkgName == "main" {
			inventory.CLIs = append(inventory.CLIs, info)
		} else {
			inventory.Packages = append(inventory.Packages, info)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return inventory, nil
}
