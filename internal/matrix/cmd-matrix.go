package matrix

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

// RunOptions holds CLI flags for running a matrix command.
type RunOptions struct {
	Dimension []string `flag:"--dimension|-d,A dimension name and values to run the command for. Eg -d GOARCH=amd64,arm64 -d GOOS=linux"`
}

// dimension represents a single matrix dimension with a name and possible values.
type dimension struct {
	Name   string
	Values []string
}

// executeOneCommand runs a single command with the current environment variables set for the matrix cell.
func executeOneCommand(ctx context.Context, _ *RunOptions, args []string) error {
	if len(args) == 0 {
		slog.WarnContext(ctx, "No command provided, skipping execution")
		return nil
	}

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

	// Log and run the command, copying stdout and stderr.
	slog.InfoContext(ctx, "  Running", "command", strings.Join(quotedCmdArgs, " "))
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
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

	slog.InfoContext(ctx, "Matrix:")

	// Initialize indexes for iterating over all combinations.
	indexes := make([]int, len(dimensions))

	for {
		// Set environment variables for each cell in the current combination.
		varstring := strings.Builder{}
		for dim, i := range indexes {
			os.Setenv(dimensions[dim].Name, dimensions[dim].Values[i])
			varstring.WriteString(fmt.Sprintf("%s=%q  ", dimensions[dim].Name, dimensions[dim].Values[i]))
		}
		slog.InfoContext(ctx, fmt.Sprintf("  Setting environment: %s", varstring.String()))

		// Run the command for the current combination.
		err := executeOneCommand(ctx, option, args)
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
	}
}
