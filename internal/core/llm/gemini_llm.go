package llm

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"github.com/markdave123-py/Contexta/internal/infra"
)

type GeminiLLM struct {
	client    *genai.Client
	modelName string
}

func NewGeminiLLM(ctx context.Context, apiKey, modelName string) (*GeminiLLM, error) {
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	cl, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}
	return &GeminiLLM{client: cl, modelName: modelName}, nil
}

func (g *GeminiLLM) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}

func (g *GeminiLLM) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	m := g.client.GenerativeModel(g.modelName)
	if systemPrompt != "" {
		m.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(systemPrompt)},
		}
	}

	resp, err := m.GenerateContent(ctx, genai.Text(userPrompt))
	if err != nil {
		return "", fmt.Errorf("gemini generate: %w", err)
	}
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return "", nil
	}

	var b strings.Builder
	for _, p := range resp.Candidates[0].Content.Parts {
		if t, ok := p.(genai.Text); ok {
			b.WriteString(string(t))
		}
	}
	return b.String(), nil
}

var _ infra.LLMProvider = (*GeminiLLM)(nil)
