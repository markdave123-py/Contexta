package ingestion_engine

import (
	"bytes"
	"context"
	"log"
	"strings"

	"code.sajari.com/docconv" // Using the corrected module path
	"github.com/markdave123-py/Contexta/internal/core"
	"golang.org/x/sync/errgroup"
)

var _ core.DocumentExtractor = (*DocconvExtractor)(nil)

func NewDocconvExtractor(useReadability bool) *DocconvExtractor {
	return &DocconvExtractor{useReadability: useReadability}
}

// ExtractText uses docconv to extract text from the given reader based on content type.
// It writes the extracted text as fragments to the output channel.
func (e *DocconvExtractor) ExtractText(ctx context.Context, g *errgroup.Group, r []byte, contentType string) (<-chan string, error) {
	out := make(chan string, 32) // Buffered channel for fragments

	reader := bytes.NewReader(r)

	go func() {
		defer close(out)

		res, err := docconv.Convert(reader, contentType, e.useReadability)
		if err != nil {
			log.Printf("docconv: extraction failed for content type '%s' (OCR: %t): %v", contentType, e.useReadability, err)
			return 
		}

		if err := ctx.Err(); err != nil {
			println("context canlled after extraction")
			return
		}

		text := res.Body

		if text == "" {
			log.Printf("docconv: extracted empty text for content type '%s'", contentType)
			return
		}

		// Split the extracted text into fragments
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			if line = strings.TrimSpace(line); line == "" {
				continue
			}
			select {
			case out <- line:
			case <-ctx.Done():

				return // Context cancelled, stop processing
			}
		}
	}()

	return out, nil
}
