package matrix

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"strings"

	"github.com/davidjspooner/go-text-cli/pkg/ansi/layout"
)

// RunOptions holds CLI flags for running a matrix command.
type RunOptions struct {
	Dimension  []string `flag:"--dimension|-d,A dimension name and values to run the command for. Eg -d GOARCH=amd64,arm64 -d GOOS=linux"`
	Shell      string   `flag:"--shell|-s,The shell to use for running commands. Defaults to bash if not set."`
	AllowColor bool     `flag:"--allow-color,-c,Allow color output in the command execution."`
	WrapWidth  int      `flag:"--wrap-width,-w,The maximum width for wrapping output. Defaults to maxint if not set."`
}

// dimension represents a single matrix dimension with a name and possible values.
type dimension struct {
	Name   string
	Values []string
}

// executeOneCommand runs a single command with the current environment variables set for the matrix cell.
func executeOneCommand(ctx context.Context, options *RunOptions, stdin []byte, args []string) error {
	// Prepare command arguments, expanding environment variables.
	cmdArgs := make([]string, len(args))
	quotedCmdArgs := make([]string, len(cmdArgs))
	for i, arg := range args {
		arg = os.ExpandEnv(arg)
		cmdArgs[i] = arg
		// Quote arguments with spaces or quotes for logging.
		if strings.Contains(arg, " ") || strings.Contains(arg, "\t") || strings.Contains(arg, "\"") {
			quotedCmdArgs[i] = fmt.Sprintf("%q", arg)
		} else {
			quotedCmdArgs[i] = arg
		}
	}
	stdinStr := os.ExpandEnv(string(stdin))
	stdinStream := io.NopCloser(strings.NewReader(stdinStr))

	if options.WrapWidth == 0 {
		options.WrapWidth = math.MaxInt // Default to maxint if not set.
	}
	wrapSpec := &layout.WrapSpec{
		MaxWidth: math.MaxInt64, // Set a fixed width for the output.
		Prefix:   "    ",        // Prefix each line with spaces for better readability.
		Width:    options.WrapWidth,
	}
	if options.AllowColor {
		wrapSpec.Color = layout.AllowColor
	}

	// Log and run the command, copying stdout and stderr.
	slog.InfoContext(ctx, "  Running", "command", strings.Join(quotedCmdArgs, " "))
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdin = stdinStream
	cmd.Stdout = layout.NewWrapStream(os.Stdout, wrapSpec) // Prefix stdout with spaces for better readability.
	cmd.Stderr = layout.NewWrapStream(os.Stderr, wrapSpec) // Prefix stderr with spaces for better readability.
	// Execute the command and handle errors.
	if err := cmd.Run(); err != nil {
		// Output command with quotes when needed.
		return fmt.Errorf("command '%s' failed: %w", strings.Join(quotedCmdArgs, " "), err)
	}
	fmt.Println("")
	return nil
}

// doMatrixExecute runs the provided command for every combination of matrix dimensions.
func doMatrixExecute(ctx context.Context, option *RunOptions, args []string) error {
	// Parse dimensions from CLI options.
	dimensions := make([]dimension, 0, len(option.Dimension))
	slog.DebugContext(ctx, "Matrix dimensions", "dimensions", option.Dimension)
	for _, dim := range option.Dimension {
		parts := strings.SplitN(dim, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid dimension format, expected 'name=value1,value2,...valueN,', got '%s'", dim)
		}
		name := parts[0]
		values := strings.Split(parts[1], ",")
		if len(values) == 0 {
			return fmt.Errorf("dimension '%s' must have at least one value", name)
		}
		dimensions = append(dimensions, dimension{
			Name:   name,
			Values: values,
		})
	}

	// Initialize indexes for iterating over all combinations.
	indexes := make([]int, len(dimensions))

	var stdin []byte
	var err error

	if len(args) == 0 || args[0] == "-" {
		if option.Shell == "" {
			option.Shell = "bash"
		}
		slog.InfoContext(ctx, fmt.Sprintf("No command provided, assuming %q with stdin", option.Shell))
		args = strings.Split(option.Shell, " ")
		stdin, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("error reading stdin: %w", err)
		}
	}
	slog.InfoContext(ctx, "Matrix:")

	for {
		// Set environment variables for each cell in the current combination.
		varstring := strings.Builder{}
		for dim, i := range indexes {
			os.Setenv(dimensions[dim].Name, dimensions[dim].Values[i])
			varstring.WriteString(fmt.Sprintf("%s=%q  ", dimensions[dim].Name, dimensions[dim].Values[i]))
		}
		slog.InfoContext(ctx, fmt.Sprintf("  Setting environment: %s", varstring.String()))

		// Run the command for the current combination.
		err := executeOneCommand(ctx, option, stdin, args)
		if err != nil {
			return fmt.Errorf("error running command: %w", err)
		}

		// Increment indexes to move to the next combination.
		for i := len(indexes) - 1; i >= 0; i-- {
			indexes[i]++
			if indexes[i] < len(dimensions[i].Values) {
				break
			}
			indexes[i] = 0
			// If we've wrapped around all indexes, we're done.
			if i == 0 {
				return nil // All combinations have been processed
			}
		}
		if len(indexes) == 0 {
			slog.WarnContext(ctx, "No dimensions provided so only itterating once")
			return nil
		}
	}
}
