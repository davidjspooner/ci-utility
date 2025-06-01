package matrix

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

type MatrixRunOptions struct {
	Dimension []string `flag:"--dimension|-d,A dimension name and values to run the command for. Eg -d GOARCH=amd64,arm64 -d GOOS=linux"`
}

type dimension struct {
	Name   string
	Values []string
}

func executeOneCommand(ctx context.Context, option *MatrixRunOptions, args []string) error {
	if len(args) == 0 {
		slog.WarnContext(ctx, "No command provided, skipping execution")
		return nil
	}

	// create command to run (expanding environment variables)
	cmdArgs := make([]string, len(args))
	quotedCmdArgs := make([]string, len(cmdArgs))
	for i, arg := range args {
		arg = os.ExpandEnv(arg)
		cmdArgs[i] = arg
		if strings.Contains(arg, " ") || strings.Contains(arg, "\t") || strings.Contains(arg, "\"") {
			quotedCmdArgs[i] = fmt.Sprintf("%q", arg)
		} else {
			quotedCmdArgs[i] = arg
		}
	}

	//run the command , copying stdout and stderr
	slog.InfoContext(ctx, "Running", "command", strings.Join(quotedCmdArgs, " "))
	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		//output command with quotes when needed
		return fmt.Errorf("command '%s' failed: %w", strings.Join(quotedCmdArgs, " "), err)
	}
	return nil
}

func doMatrixExecute(ctx context.Context, option *MatrixRunOptions, args []string) error {
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

	indexes := make([]int, len(dimensions))

	for {
		// Set environment variables for each cell
		for dim, i := range indexes {
			os.Setenv(dimensions[dim].Name, dimensions[dim].Values[i])
		}

		err := executeOneCommand(ctx, option, args)
		if err != nil {
			return fmt.Errorf("error running command: %w", err)
		}

		//increment indexes
		for i := len(indexes) - 1; i >= 0; i-- {
			indexes[i]++
			if indexes[i] < len(dimensions[i].Values) {
				break
			}
			indexes[i] = 0
			if i == 0 {
				return nil // All combinations have been processed
			}
		}
	}
}
