package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"github.com/markdave123-py/Contexta/internal/infra"
)

type GeminiEmbedder struct {
	client    *genai.Client
	modelName string
}

func NewGeminiEmbedder(ctx context.Context, apiKey, modelName string) (*GeminiEmbedder, error) {
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	cl, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	if modelName == "" {
		modelName = "gemini-embedding-001"
	}
	return &GeminiEmbedder{client: cl, modelName: modelName}, nil
}


func (g *GeminiEmbedder) Close() error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}

// EmbedTexts batches all texts in one request via EmbeddingBatch.
func (g *GeminiEmbedder) EmbedTexts(ctx context.Context, texts []string, dim int) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	em := g.client.EmbeddingModel(g.modelName)

	batch := em.NewBatch()
	for _, t := range texts {
		batch.AddContent(genai.Text(t))
	}

	resp, err := em.BatchEmbedContents(ctx, batch)
	if err != nil {
		return nil, fmt.Errorf("gemini batch embed: %w", err)
	}

	out := make([][]float32, 0, len(resp.Embeddings))
	for _, e := range resp.Embeddings {
		out = append(out, e.Values)
	}
	return out, nil
}

var _ infra.EmbeddingProvider = (*GeminiEmbedder)(nil)
