package llm

import (
	"context"
	"fmt"

	"github.com/davidjspooner/go-llm-client/pkg/llmclient"
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

type LLMOptions struct {
	// Add any options specific to the review command here
	Config       string `flag:"--config,config file for LLM model to use"`
	SystemPrompt string `flag:"--system,system prompt file for LLM"`
	UserPrompt   string `flag:"--user,user prompt file for LLM"`
	OutputFile   string `flag:"--output,output file for LLM response"`
}

// Commands returns the list of git-related CLI commands for the application.
func AddCommandsTo(parent cmd.Command) error {
	llm := cmd.NewCommand(
		"llm",
		"LLM content generation",
		runLLM,
		&LLMOptions{},
	)

	// Add subcommands to the root git command.
	err := parent.SubCommands().Add(llm)
	if err != nil {
		return fmt.Errorf("failed to add LLM command: %w", err)
	}

	return nil
}

func runLLM(ctx context.Context, options *LLMOptions, args []string) error {
	clientConfig, err := llmclient.LoadConfig(options.Config)
	if err != nil {
		return fmt.Errorf("failed to load LLM config: %w", err)
	}
	client, err := clientConfig.CreateClient()
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}

	req := llmclient.Request{}

	system, err := cmd.ReadFileOrSpecial(options.SystemPrompt)
	if err != nil {
		return fmt.Errorf("failed to read system prompt: %w", err)
	}
	if system != "" {
		req.Messages = append(req.Messages, llmclient.Message{
			Role:    llmclient.RoleSystem,
			Content: system,
		})
	}
	user, err := cmd.ReadFileOrSpecial(options.UserPrompt)
	if err != nil {
		return fmt.Errorf("failed to read user prompt: %w", err)
	}
	if user != "" {
		req.Messages = append(req.Messages, llmclient.Message{
			Role:    llmclient.RoleUser,
			Content: user,
		})
	}

	for _, arg := range args {
		req.Messages = append(req.Messages, llmclient.Message{
			Role:    llmclient.RoleUser,
			Content: arg,
		})
	}

	if len(req.Messages) == 0 {
		return fmt.Errorf("no messages provided for LLM request")
	}

	resp, err := client.Chat(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to get LLM response: %w", err)
	}
	for _, choice := range resp.Choices {
		fmt.Println("Response:", choice.Message.Content)
	}
	return nil
}
