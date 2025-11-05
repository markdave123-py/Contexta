package ingestion_engine

import (
	"github.com/markdave123-py/Contexta/internal/core"
)

// IngestConfig tunes the streaming pipeline.
//
// TargetTokens:   approximate tokens per chunk (e.g., 500).
// OverlapTokens:  token overlap between consecutive chunks for context bleed (e.g., 50).
// BatchSize:      how many chunks to embed/write in one batch (e.g., 32).
// MaxFragmentLen: soft upper bound for individual fragments coming from the extractor.
// EmbedDim:       embedding dimension (use 0 to let model default apply; set to 768 if you want IVF on pgvector).
type IngestConfig struct {
	TargetTokens   int
	OverlapTokens  int
	BatchSize      int
	EmbedDim       int
}

// chunk is the internal representation passed through the pipeline.
//
// Pos:      stable, zero-based position of the chunk inside the document.
// Text:     chunk content (built from one or more fragments).
// TokenCnt: approximate token count (used for batching and overlap math).
type chunk struct {
	Pos      int
	Text     string
	TokenCnt int
}

// DocumentIngestor orchestrates the background ingestion pipeline:
//
// db:        persistence for document and chunks.
// obj:       object storage for streaming large files.
// embedder:  embedding provider (Gemini/OpenAI/etc).
// cfg:       runtime tuning knobs for the pipeline.
// jobs:      in-memory queue of document IDs to process (easy to swap with Kafka later).
type DocumentIngestor struct {
	db       core.DbClient
	obj      core.ObjectClient
	embedder core.EmbeddingProvider
	extrator core.DocumentExtractor
	cfg      *IngestConfig
	jobs     chan string
}

// DocumentExtractor implements core.DocumentExtractor using sajari/docconv.
type DocconvExtractor struct {
	useReadability bool
}
