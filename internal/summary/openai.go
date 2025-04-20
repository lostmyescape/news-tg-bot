package summary

import (
	"context"
	"github.com/lostmyescape/news-tg-bot/logger"
	"github.com/sashabaranov/go-openai"
	"strings"
	"sync"
)

type OpenAISummarizer struct {
	client  *openai.Client
	prompt  string
	enabled bool
	mu      sync.Mutex
}

func NewOpenAiSummarizer(apiKey string, prompt string) *OpenAISummarizer {
	s := &OpenAISummarizer{
		client:  openai.NewClient(apiKey),
		prompt:  prompt,
		enabled: apiKey != "",
	}

	logger.Log.Infow("openai summarizer created", "enabled", s.enabled)

	return s
}

func (s *OpenAISummarizer) Summarize(ctx context.Context, text string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.enabled {
		return "", nil
	}

	request := openai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: s.prompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: text,
			},
		},
		MaxTokens:   256,
		Temperature: 0.7,
		TopP:        1,
	}

	resp, err := s.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", err
	}

	rawSummary := strings.TrimSpace(resp.Choices[0].Message.Content)
	if !strings.HasSuffix(rawSummary, ".") {
		rawSummary += "."
	}

	return rawSummary, nil
}
