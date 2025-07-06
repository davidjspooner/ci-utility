package golang

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

// ReviewOptions holds options for the review command.
type ReviewOptions struct {
	// Add any options specific to the review command here
	Report      string `flag:"--report,Path to save the review report"`
	TargetScore int    `flag:"--target-score,Target score for the review"`
}

// reviewCommand defines the CLI command for code review.
var reviewCommand = cmd.NewCommand(
	"review",
	"Review project and optionally generate a review.yaml file",
	func(ctx context.Context, options *ReviewOptions, args []string) error {
		meta := Scope{}
		for _, arg := range args {
			roots, err := filepath.Glob(arg)
			if err != nil {
				return fmt.Errorf("failed to glob %s: %w", arg, err)
			}
			meta.RootPaths = append(meta.RootPaths, roots...)
		}

		r := GoReview{}
		results, err := r.Run(ctx, &meta, options)
		if err != nil {
			return err
		}

		slices.SortFunc(results, func(a, b *Result) int {
			return a.Score - b.Score
		})

		for _, result := range results {
			for _, issue := range result.Issues {
				fmt.Printf("  - %s[%d]: %s %s\n", issue.Filename, issue.Line, issue.Type, issue.Message)
			}
		}
		fmt.Printf("Total issues found: %d\n", len(r.issues))
		return nil
	},
	&ReviewOptions{
		TargetScore: 100,
	},
)

// GoReview analyzes go code quality (as defined by me).
type GoReview struct {
	issues []*Issue
}

var _ Category = (*GoReview)(nil)

// Name returns the name of the GoReview category.
func (c *GoReview) Name() string { return "go_code_structure" }

// Run executes the GoReview checks on the provided meta and options.
func (c *GoReview) Run(ctx context.Context, meta *Scope, options *ReviewOptions) ([]*Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var packages goPackageInventory
	for _, rootPath := range meta.RootPaths {
		// Add packages found at each root path.
		err := packages.addPackages(rootPath)
		if err != nil {
			return nil, err
		}
	}
	for _, p := range packages {
		// Scan files in each package for issues.
		p.scanFiles(
			c.checkTodosfunc,
			c.checkExports,
			c.checkLargeFunctions,
			c.checkCommentRatio,
			c.checkMustVariants,
		)
	}
	result := Result{Name: c.Name(), Score: 0, Issues: nil}
	for _, issue := range c.issues {
		// Add each found issue to the result.
		result.Issues = append(result.Issues, issue)
	}
	//result.Summerize()

	return []*Result{&result}, nil
}

// checkTodosfunc checks for TODOs and not implemented hints in comments and code.
func (c *GoReview) checkTodosfunc(pi *goPackageInfo, filename string, fset *token.FileSet, file *ast.File) error {
	hints := []string{"todo", "not implemented", "notimplemented"}

	// Helper to check if a string contains any hint (case-insensitive)
	containsHint := func(s string) (string, bool) {
		ls := strings.ToLower(s)
		for _, h := range hints {
			if strings.Contains(ls, h) {
				return h, true
			}
		}
		return "", false
	}

	// Check comments for TODOs.
	for _, cg := range file.Comments {
		for _, cmt := range cg.List {
			if _, ok := containsHint(cmt.Text); ok {
				position := fset.Position(cmt.Pos())
				c.addIssue(pi, filename, position.Line, "todo", cmt.Text)
			}
		}
	}

	// Walk function bodies for string literals and identifiers.
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.BasicLit:
			if x.Kind == token.STRING {
				if _, ok := containsHint(x.Value); ok {
					position := fset.Position(x.Pos())
					c.addIssue(pi, filename, position.Line, "todo", x.Value)
				}
			}
		case *ast.Ident:
			if _, ok := containsHint(x.Name); ok {
				position := fset.Position(x.Pos())
				c.addIssue(pi, filename, position.Line, "todo", x.Name)
			}
		}
		return true
	})

	return nil
}

// checkExports checks for missing documentation and naming issues on exported identifiers.
func (c *GoReview) checkExports(pi *goPackageInfo, filename string, fset *token.FileSet, file *ast.File) error {
	// Walk the AST to find exported declarations and check for documentation and naming issues.
	ast.Inspect(file, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.GenDecl:
			// Handle type, const, and var declarations.
			if decl.Tok == token.TYPE || decl.Tok == token.CONST || decl.Tok == token.VAR {
				for _, spec := range decl.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						// Check exported types.
						if ast.IsExported(s.Name.Name) {
							// Check for missing doc comment.
							if decl.Doc == nil {
								pos := fset.Position(s.Pos())
								c.addIssue(pi, filename, pos.Line, "missing godoc", "Exported type '"+s.Name.Name+"' lacks a comment")
							}
							// Check for redundant package name prefix.
							if pi.Type == GoPackageTypePackage && strings.HasPrefix(strings.ToLower(s.Name.Name), strings.ToLower(file.Name.Name)) && (len(s.Name.Name) > len(file.Name.Name)) {
								pos := fset.Position(s.Pos())
								c.addIssue(pi, filename, pos.Line, "naming", "Exported type '"+s.Name.Name+"' redundantly starts with package name")
							}
						}
					case *ast.ValueSpec:
						// Check exported values (consts/vars).
						for _, name := range s.Names {
							if ast.IsExported(name.Name) {
								// Check for missing doc comment on value or the group.
								if decl.Doc == nil && s.Doc == nil {
									pos := fset.Position(name.Pos())
									c.addIssue(pi, filename, pos.Line, "missing godoc", "Exported value '"+name.Name+"' lacks a comment")
								}
								// Check for redundant package name prefix.
								if pi.Type == GoPackageTypePackage && strings.HasPrefix(strings.ToLower(name.Name), strings.ToLower(file.Name.Name)) && (len(name.Name) > len(file.Name.Name)) {
									pos := fset.Position(name.Pos())
									c.addIssue(pi, filename, pos.Line, "naming", "Exported value '"+name.Name+"' redundantly starts with package name")
								}
							}
						}
					}
				}
			}
		case *ast.FuncDecl:
			// Check exported functions (not methods).
			if decl.Recv == nil && ast.IsExported(decl.Name.Name) {
				// Check for missing doc comment.
				if decl.Doc == nil {
					pos := fset.Position(decl.Pos())
					c.addIssue(pi, filename, pos.Line, "missing godoc", "Exported function '"+decl.Name.Name+"' lacks a comment")
				}
				// Check for redundant package name prefix.
				if pi.Type == GoPackageTypePackage && strings.HasPrefix(strings.ToLower(decl.Name.Name), strings.ToLower(file.Name.Name)) && (len(decl.Name.Name) > len(file.Name.Name)) {
					pos := fset.Position(decl.Pos())
					c.addIssue(pi, filename, pos.Line, "naming", "Exported function '"+decl.Name.Name+"' redundantly starts with package name")
				}
			}
		}
		// Continue walking the AST.
		return true
	})
	return nil
}

// checkLargeFunctions checks for functions that are too large.
func (c *GoReview) checkLargeFunctions(pi *goPackageInfo, filename string, fset *token.FileSet, file *ast.File) error {
	const maxLines = 60

	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			start := fset.Position(fn.Pos()).Line
			end := fset.Position(fn.End()).Line
			if end-start > maxLines {
				c.addIssue(pi, filename, start, "size", "Function '"+fn.Name.Name+"' is too large ("+strconv.Itoa(end-start)+" lines)")
			}
		}
		return true
	})
	return nil
}

// checkCommentRatio checks if functions have a sufficient comment-to-code ratio.
func (c *GoReview) checkCommentRatio(pi *goPackageInfo, filename string, fset *token.FileSet, file *ast.File) error {
	if pi.Type != GoPackageTypePackage {
		return nil // only check main packages
	}
	ast.Inspect(file, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			return true
		}

		start := fset.Position(fn.Pos()).Line
		end := fset.Position(fn.End()).Line
		totalLines := end - start
		if totalLines < 20 { //dont worry about small functions
			return true
		}

		commentLines := 0
		for _, cg := range file.Comments {
			pos := fset.Position(cg.Pos()).Line
			if pos >= start && pos <= end {
				commentLines += len(strings.Split(cg.Text(), "\n"))
			}
		}

		ratio := float64(commentLines) / float64(totalLines)
		if ratio < 0.1 {
			c.addIssue(pi, filename, start, "comment-ratio", "Function '"+fn.Name.Name+"' has low comment ratio: "+fmt.Sprintf("%.2f", ratio))
		}

		return true
	})
	return nil
}

// checkMustVariants checks for the presence of Must* variants for exported New/Parse functions that return error.
func (c *GoReview) checkMustVariants(pi *goPackageInfo, filename string, fset *token.FileSet, file *ast.File) error {
	// Collect exported New/Parse function names
	funcs := make(map[string]*ast.FuncDecl)
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Recv == nil && ast.IsExported(fn.Name.Name) {
			if strings.HasPrefix(fn.Name.Name, "New") || strings.HasPrefix(fn.Name.Name, "Parse") {
				funcs[fn.Name.Name] = fn
			}
		}
		return true
	})

	// Check for corresponding Must variants
	for name, fn := range funcs {
		// Only require Must variant if the function returns an error
		hasErrorReturn := false
		if fn.Type.Results != nil {
			for _, result := range fn.Type.Results.List {
				// Check if the result type is error
				if ident, ok := result.Type.(*ast.Ident); ok && ident.Name == "error" {
					hasErrorReturn = true
					break
				}
			}
		}
		if !hasErrorReturn {
			continue
		}

		mustName := "Must" + name
		found := false
		for _, f := range file.Decls {
			if fd, ok := f.(*ast.FuncDecl); ok && fd.Recv == nil && fd.Name.Name == mustName {
				found = true
				break
			}
		}
		if !found {
			pos := fset.Position(fn.Pos())
			c.addIssue(pi, filename, pos.Line, "must-variant", fmt.Sprintf("Exported function '%s' has no corresponding '%s'", name, mustName))
		}
	}
	return nil
}

// addIssue adds a new issue to the GoReview issues list.
func (c *GoReview) addIssue(pi *goPackageInfo, filename string, lineNumber int, issueType string, message string) {
	//create a new issue and add it to the package's issues
	issue := &Issue{
		Package:  pi.Name,
		Filename: filename,
		Line:     lineNumber,
		Type:     issueType,
		Message:  message,
		Weight:   1, // Initialize count to 1 for the first occurrence
	}
	c.issues = append(c.issues, issue)
}
