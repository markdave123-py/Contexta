package core

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// ExtractedText represents the result of text extraction, potentially with metadata.
type ExtractedText struct {
	Text     string
	Metadata map[string]string
}

// DocumentExtractor defines the interface for extracting text from various document types.
type DocumentExtractor interface {
	// ExtractText takes an io.Reader and content type, and returns a channel of extracted text fragments.
	// The `contentType` hint helps the extractor choose the right parsing strategy.
	ExtractText(ctx context.Context, g *errgroup.Group, r []byte, contentType string) (<-chan string, error)
}
