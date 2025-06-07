package golang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
)

// goPackageType represents the type of a Go package (CLI, package, or test).
type goPackageType string

const (
	// GoPackageTypeCLI indicates a main CLI package.
	GoPackageTypeCLI goPackageType = "cli"

	// GoPackageTypePackage indicates a standard Go package.
	GoPackageTypePackage goPackageType = "package"

	// GoPackageTypeTest indicates a test package.
	GoPackageTypeTest goPackageType = "test"
)

// goPackageInfo holds information about a Go package.
type goPackageInfo struct {
	DirPath string
	Name    string
	Type    goPackageType
}

// goPackageInventory is a slice of goPackageInfo representing discovered packages.
type goPackageInventory []goPackageInfo

// contains checks if the inventory contains a package with the same DirPath and Name.
func (inventory *goPackageInventory) contains(pi *goPackageInfo) bool {
	for _, p := range *inventory {
		// Check for matching directory and package name.
		if p.DirPath == pi.DirPath && p.Name == pi.Name {
			return true
		}
	}
	return false
}

// addPackages walks a project root and builds categorized package inventory.
func (inventory *goPackageInventory) addPackages(root string) error {
	// WalkDir traverses the directory tree rooted at root.
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		// Skip on error, directories, non-Go files, or hidden files.
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		// Parse the Go file to get the package name.
		fset := token.NewFileSet()
		f, parseErr := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if parseErr != nil || f == nil {
			return nil
		}

		dir := filepath.Dir(path)
		pkgName := f.Name.Name

		info := goPackageInfo{DirPath: dir, Name: pkgName}

		// Avoid duplicate entries in the inventory.
		if inventory.contains(&info) {
			return nil
		}

		// Classify the package type.
		if strings.HasSuffix(path, "_test.go") {
			info.Type = GoPackageTypeTest
			*inventory = append(*inventory, info)
			return nil
		}

		if pkgName == "main" {
			info.Type = GoPackageTypeCLI
			*inventory = append(*inventory, info)
		} else {
			info.Type = GoPackageTypePackage
			*inventory = append(*inventory, info)
		}
		return nil
	})

	// Return any error encountered during walking.
	if err != nil {
		return err
	}
	return nil
}

// fileCheckerFunc is a function type for checking files in a package.
type fileCheckerFunc func(pi *goPackageInfo, filename string, fset *token.FileSet, f *ast.File) error

// scanFiles enumerates Go files in the package directory and applies checker functions.
func (pi *goPackageInfo) scanFiles(funcs ...fileCheckerFunc) error {
	// Enumerate files in the package directory.
	files, err := filepath.Glob(filepath.Join(pi.DirPath, "*.go"))
	if err != nil {
		return err // handle error appropriately
	}
	for _, file := range files {
		// Generate AST for file.
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, file, nil, parser.ParseComments|parser.AllErrors)
		if err != nil {
			return err // handle error appropriately
		}
		// Apply each checker function to the file.
		for _, fn := range funcs {
			if err := fn(pi, file, fset, f); err != nil {
				return err // handle error appropriately
			}
		}
	}
	return nil // all files processed successfully
}
