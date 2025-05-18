package template

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	htmlTemplate "html/template"
	textTemplate "text/template"

	"github.com/davidjspooner/go-text/pkg/cmd"
)

type expandOptions struct {
	Type   string `flag:"--format,Type of template to expand (go/text, go/html, etc.)"`
	Target string `flag:"--target,Target file/directory to expanded into  ( use trailing / for directory )"`
}

var templateFunctions = map[string]any{
	"env": func(key string) string {
		return os.Getenv(key)
	},
	"file": func(path string) (string, error) {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", path, err)
		}
		return string(data), nil
	},
}

func expandTemplate(ctx context.Context, options *expandOptions, args []string) error {

	target := options.Target
	if target == "" {
		return fmt.Errorf("--target is required")
	}
	isTargetDir := target[len(target)-1] == '/'
	if len(args) == 0 {
		return fmt.Errorf("no files specified")
	}
	if !isTargetDir && len(args) > 1 {
		return fmt.Errorf("multiple files specified, but target is not a directory")
	}
	if isTargetDir {
		err := os.MkdirAll(target, 0755)
		if err != nil {
			return fmt.Errorf("failed to create target directory %s: %w", target, err)
		}
	}

	for _, arg := range args {
		_, err := os.Stat(arg)
		if err != nil {
			return fmt.Errorf("failed to stat file %s: %w", arg, err)
		}

		targetName := target
		if isTargetDir {
			base := path.Base(arg)
			base = strings.Replace(base, ".tmpl", "", -1)
			targetName = target + base
		}
		err = expandTemplateFile(arg, targetName, options.Type)
		if err != nil {
			return fmt.Errorf("failed to expand template %s: %w", arg, err)
		}
	}
	return nil
}

func expandTemplateFile(source, target, templateType string) error {
	f, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("failed to open template file %s: %w", source, err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", source, err)
	}

	targetFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("failed to create target file %s: %w", target, err)
	}
	defer targetFile.Close()

	switch templateType {
	case "go/text":
		tmpl, err := textTemplate.New("template").Funcs(templateFunctions).Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse text template %s: %w", source, err)
		}
		err = tmpl.Execute(targetFile, nil)
		if err != nil {
			return fmt.Errorf("failed to expand text template %s: %w", source, err)
		}
	case "go/html":
		tmpl, err := htmlTemplate.New("template").Funcs(templateFunctions).Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse HTML template %s: %w", source, err)
		}
		err = tmpl.Execute(targetFile, nil)
		if err != nil {
			return fmt.Errorf("failed to expand HTML template %s: %w", source, err)
		}
	default:
		return fmt.Errorf("unsupported template type: %s", templateType)
	}

	return nil
}

func templateManPage(ctx context.Context, options *cmd.NoopOptions, args []string) error {
	print("# TODO")
	return nil
}
