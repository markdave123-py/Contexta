package core

import "context"

type EmbeddingProvider interface {
	EmbedTexts(ctx context.Context, texts []string) ([][]float32, error)
}

type LLMProvider interface {
	Generate(ctx context.Context, systemPrompt string, userPrompt string) (string, error)
}
