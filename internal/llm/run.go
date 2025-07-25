package llm

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/davidjspooner/go-llm-client/pkg/llmclient"
	"github.com/davidjspooner/go-text-cli/pkg/cmd"
)

type CustomOptions struct {
	// Add any options specific to the review command here
	SystemPrompt string `flag:"--system,system prompt file for LLM"`
}

func executeCustomCommand(ctx context.Context, options *CustomOptions, args []string) error {

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

	req := llmclient.Request{}

	system, err := cmd.ReadFileOrSpecial(options.SystemPrompt)
	if err != nil {
		return fmt.Errorf("failed to read system prompt: %w", err)
	}
	if system != nil {
		req.Messages = append(req.Messages, llmclient.Message{
			Role:    llmclient.RoleSystem,
			Content: string(system),
		})
	}
	for n, arg := range args {
		if arg == "" {
			return fmt.Errorf("argument #%d is empty", n)
		}

		if arg == "-" {
			arg = "/dev/stdin"
		}
		if arg[0] == '!' { // inline text
			req.Messages = append(req.Messages, llmclient.Message{
				Role:    llmclient.RoleUser,
				Content: arg[1:],
			})
			continue
		}

		content, err := cmd.ReadFileOrSpecial(arg)
		if err != nil {
			return fmt.Errorf("failed to read argument %d (%s): %w", n, arg, err)
		}
		mimetype := http.DetectContentType(content)
		if strings.HasPrefix(mimetype, "text/") || strings.HasPrefix(mimetype, "application/json") || strings.HasPrefix(mimetype, "application/x-yaml") {
			req.Messages.AppendUserTextFile(arg, mimetype, content)
			continue
		}
		req.Messages.AppendUserBase64File(arg, mimetype, content)
	}

	if len(req.Messages) == 0 {
		return fmt.Errorf("no messages provided for LLM request")
	}

	resp, err := client.Chat(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to get LLM response: %w", err)
	}
	if llmOptions.OutputFile == "-" || llmOptions.OutputFile == "" {
		for _, choice := range resp.Choices {
			fmt.Println("Response:", choice.Message.Content)
		}
	} else {
		f, err := os.Open(llmOptions.OutputFile)
		if err != nil {
			return fmt.Errorf("failed to open output file %s: %w", llmOptions.OutputFile, err)
		}
		defer f.Close()
		for n, choice := range resp.Choices {
			if n > 0 {
				f.Write([]byte{'\n'}) // Separate choices with a newline
			}
			_, err := f.Write([]byte(choice.Message.Content))
			if err != nil {
				return fmt.Errorf("failed to write to output file %s: %w", llmOptions.OutputFile, err)
			}
		}
	}
	return nil
}
