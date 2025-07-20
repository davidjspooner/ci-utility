package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/davidjspooner/go-llm-client/pkg/llmclient"
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

var promptForInterpetingHelpOutput = `
You are a system that analyzes CLI help output and recursively builds a structured representation of a command-line tool in a format suitable for OpenAI function calling.

Given:
1. The exact CLI command that was executed.
2. The textual output returned from that command.

Your task is to:
1. Determine which help flag is used ("--help", "-h", "-?", etc.) based on the command that was run.
2. Identify and return a list of **leaf commands**. These are commands that appear to be directly callable (i.e., perform actions and have no subcommands). For each one, generate a JSON function definition in the OpenAI function-calling format:
   - "name": full command string, with spaces replaced by underscores or hyphens
   - "description": from the help output
   - "parameters": JSON Schema object containing:
	 - Positional arguments
	 - Flags (including global flags, which should be repeated in every leaf)
	 - Each property should include: "type", "description", and "enum" if options are shown
	 - Include a "required" list if any are explicitly required

3. Identify and return **expandable commands** — commands that likely contain additional subcommands. For each one, return a string that includes:
   - The full command to execute
   - The appropriate help flag inferred from the original input ("--help", "-h", or "-?")

Return a JSON object with two keys:
- "leaf_commands": list of OpenAI function definitions
- "expandable_commands": list of strings to run to extract further help

Use the following placeholders:

----BEGIN COMMAND----
%[1]s
----END COMMAND----
----BEGIN OUTPUT----
%[2]s
----END OUTPUT----
`

type ModifyToolsDefOptions struct {
	// Add any options specific to the review command here
	Overwrite bool `flag:"--overwrite,overwrite existing tools definition"`
}

type Response struct {
	Tools  []llmclient.ToolDefinition `json:"leaf_commands"`
	Expand []string                   `json:"expandable_commands"`
}

type HelpExplorer struct {
	client llmclient.ChatClient
}

func (he *HelpExplorer) runCommand(_ context.Context, command string) (string, error) {
	cmdParts := strings.Split(command, " ")
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("command %q failed: %v", command, err)
	}
	return strings.TrimSpace(out.String()), nil
}

func (he *HelpExplorer) Explore(ctx context.Context, helpCommand string) (*Response, error) {

	// Execute the help command to get the output
	helpOutput, err := he.runCommand(ctx, helpCommand)
	if err != nil {
		return nil, fmt.Errorf("failed to run help command %q: %w", helpCommand, err)
	}

	req := llmclient.Request{}
	req.Messages = append(req.Messages, llmclient.Message{
		Role:    llmclient.RoleUser,
		Content: fmt.Sprintf(promptForInterpetingHelpOutput, helpCommand, helpOutput),
	})
	resp, err := he.client.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	var response Response
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &response, nil
}

func addToolsDefCommand(ctx context.Context, options *ModifyToolsDefOptions, args []string) error {

	llmOptions, err := cmd.FindOptionStruct[LLMOptions](ctx)
	if err != nil {
		return fmt.Errorf("failed to find LLM options: %w", err)
	}

	clientConfig, err := llmclient.LoadConfig(llmOptions.Config)
	if err != nil {
		return fmt.Errorf("failed to load LLM config: %w", err)
	}
	client, err := clientConfig.CreateChatClient()
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}

	he := &HelpExplorer{
		client: client,
	}

	for n := 0; n < len(args); n++ {
		arg := args[n]
		response, err := he.Explore(ctx, arg)
		if err != nil {
			return fmt.Errorf("failed to explore help for %q: %w", arg, err)
		}
		for _, tool := range response.Tools {
			// Here you would typically save the tool definition to a file or database
			// For this example, we'll just print it
			toolJSON, err := json.MarshalIndent(tool, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal tool definition: %w", err)
			}
			fmt.Println(string(toolJSON))
		}
		for _, expandCmd := range response.Expand {
			if slices.Contains(args, expandCmd) {
				continue // Skip if already in args
			}
			args = append(args, expandCmd) // Add to args for further processing
		}
	}

	return nil
}

func removeToolsDefCommand(ctx context.Context, options *ModifyToolsDefOptions, args []string) error {
	// This function is a placeholder for future implementation
	return fmt.Errorf("removeToolsDefCommand not implemented yet")
}
