package ingestorengine

import (
	"context"
	"fmt"

	"github.com/markdave123-py/Contexta/internal/models"
)

// embedAndPersist consumes chunks, embeds them in batches, and writes to DB.
// This function provides the downstream sink for the pipeline above.
//
// docID:      current document ID.
// in:         chunk stream from streamChunk.
// batchSize:  number of chunks to embed/write per batch (limits memory).
// embedDim:   model dimension (0 = default).
func (i *DocumentIngestor) embedAndPersist(
	ctx context.Context,
	docID string,
	in <-chan chunk,
	batchSize int,
	embedDim int,
) error {
	batch := make([]chunk, 0, batchSize)

	// flush embeds the current batch and inserts it into the database.
	flush := func(items []chunk) error {
		if len(items) == 0 {
			return nil
		}

		texts := make([]string, len(items))
		for idx := range items {
			texts[idx] = items[idx].Text
		}

		vecs, err := i.embedder.EmbedTexts(ctx, texts, embedDim)
		if err != nil {
			return fmt.Errorf("embed: %w", err)
		}
		if len(vecs) != len(items) {
			return fmt.Errorf("embed size mismatch: got %d want %d", len(vecs), len(items))
		}

		// 3) Map to persistence rows and write once.
		rows := make([]models.DocumentChunk, len(items))
		for k := range items {
			rows[k] = models.DocumentChunk{
				DocumentID: docID,
				Text:       items[k].Text,
				Embedding:  vecs[k],
				Position:   items[k].Pos,
				TokenCount: items[k].TokenCnt,
			}
		}
		if err := i.db.InsertDocumentChunks(ctx, rows); err != nil {
			return fmt.Errorf("insert chunks: %w", err)
		}
		return nil
	}

	// Read the stream and flush in batches.
	for c := range in {
		batch = append(batch, c)
		if len(batch) == batchSize {
			if err := flush(batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}
	// Final tail.
	if err := flush(batch); err != nil {
		return err
	}
	return nil
}
