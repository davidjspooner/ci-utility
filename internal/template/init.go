package template

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

// Commands returns the list of template-related CLI commands,
// including the 'template expand' subcommand.
func Commands() []cmd.Command {
	templateCommand := cmd.NewCommand(
		"template",
		"Template commands",
		nil,
		&cmd.NoopOptions{},
	)
	expand := cmd.NewCommand(
		"expand",
		"Expand a template file",
		expandTemplate,
		&expandOptions{},
	)
	templateCommand.SubCommands().MustAdd(expand)
	return []cmd.Command{templateCommand}
}

// expandOptions defines the command-line flags for the 'template expand' command.
// It controls the format of the template and the destination of the output.
type expandOptions struct {
	Type       string `flag:"--format,Type of template to expand (go/text, go/html, etc.)"`
	Target     string `flag:"--target,Target file/directory to expanded into  ( use trailing / for directory )"`
	InPlace    bool   `flag:"--in-place,Expand template in place (overwrites original file)"`
	ValuesYaml string `flag:"--values,Values yaml file with a map of NAME: value strings"`
}

// expandTemplate processes one or more template files based on the provided arguments and options.
// It handles validation of output paths, creates directories as needed, and expands each file.
func expandTemplate(ctx context.Context, options *expandOptions, args []string) error {

	if len(args) == 0 {
		return fmt.Errorf("no files specified")
	}
	isTargetDir := false
	if options.InPlace {
		if options.Target != "" {
			return fmt.Errorf("--target is not allowed with --in-place")
		}
	} else {
		target := options.Target
		if target == "" {
			return fmt.Errorf("--target or --inplace is required")
		}
		isTargetDir = target[len(target)-1] == '/'
		if !isTargetDir && len(args) > 1 {
			return fmt.Errorf("multiple files specified, but target is not a directory")
		}
		if isTargetDir {
			err := os.MkdirAll(target, 0755)
			if err != nil {
				return fmt.Errorf("failed to create target directory %s: %w", target, err)
			}
		}
	}
	values, err := loadValues(options.ValuesYaml)
	if err != nil {
		return fmt.Errorf("failed to load values from %s: %w", options.ValuesYaml, err)
	}

	for _, arg := range args {
		var targetName string
		if options.InPlace {
			targetName = arg
		} else if isTargetDir {
			targetName = path.Join(options.Target, path.Base(arg))
		} else {
			targetName = options.Target
		}

		_, err := os.Stat(arg)
		if err != nil {
			return fmt.Errorf("failed to stat file %s: %w", arg, err)
		}

		if isTargetDir {
			base := path.Base(arg)
			base = strings.Replace(base, ".tmpl", "", -1)
			targetName = targetName + base
		}
		err = expandTemplateFile(arg, targetName, options.Type, values)
		if err != nil {
			return fmt.Errorf("failed to expand template %s: %w", arg, err)
		}
	}
	return nil
}

// expandTemplateFile reads a template file from 'source', expands it into a temporary file,
// and then renames it to 'target'. It uses the specified 'templateType' to select the rendering engine.
func expandTemplateFile(source, target, templateType string, values *Values) error {
	f, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("failed to open template file %s: %w", source, err)
	}
	defer f.Close()

	//create a temporty file to write the expanded template to
	tempFile, err := os.CreateTemp("", "expanded_template_")
	if err != nil {
		return fmt.Errorf("failed to create temporary file for expanded template: %w", err)
	}
	err = expandTemplateStream(f, tempFile, templateType, values)
	if err != nil {
		return fmt.Errorf("failed to expand template stream %s: %w", source, err)
	}
	//close the temp file
	defer tempFile.Close()
	//move the temp file to the target file
	err = os.Rename(tempFile.Name(), target)
	return err
}

// expandTemplateStream reads template content from 'source', expands it using the appropriate engine
// based on 'templateType', and writes the result to 'target'.
func expandTemplateStream(source io.Reader, target io.Writer, templateType string, values *Values) error {

	content, err := io.ReadAll(source)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", source, err)
	}

	switch templateType {
	case "go/text":
		err := expandGoTextTemplate(string(content), target, values)
		if err != nil {
			return err
		}
	case "go/html":
		err := expandGoHTMLTemplate(string(content), target, values)
		if err != nil {
			return err
		}
	case "markdown":
		expander := MarkdownExpander{
			Lookup: func(key string) (string, error) {
				v := os.Getenv(key)
				if v == "" {
					return "", fmt.Errorf("environment variable %s not set", key)
				}
				return v, nil
			},
		}
		err = expander.Expand(content, target)
		if err != nil {
			return fmt.Errorf("failed to expand markdown template %s: %w", source, err)
		}
	default:
		return fmt.Errorf("unsupported template type: %s", templateType)
	}

	return nil
}
