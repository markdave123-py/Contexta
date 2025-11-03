package ingestorengine

import (
	"context"
	"fmt"
	"strings"

	"github.com/markdave123-py/Contexta/internal/core"
	"golang.org/x/sync/errgroup"
)

// NewDocumentIngestor constructs the ingestor with a bounded job queue (64).
func NewDocumentIngestor(db core.DbClient, obj core.ObjectClient, emb core.EmbeddingProvider, cfg IngestConfig) *DocumentIngestor {
	return &DocumentIngestor{
		db: db, obj: obj, embedder: emb, cfg: cfg,
		jobs: make(chan string, 64),
	}
}

// Start runs a single worker goroutine reading from the jobs channel.
func (i *DocumentIngestor) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case docID := <-i.jobs:
				// Best-effort processing; errors can be logged or reported.
				_ = i.processOne(ctx, docID)
			}
		}
	}()
}

// Enqueue schedules a document ID for ingestion.
// If the queue is full, this call will block until space frees up.
func (i *DocumentIngestor) Enqueue(docID string) {
	i.jobs <- docID
}

// processOne streams, chunks, embeds and persists for a single document ID.
func (i *DocumentIngestor) processOne(ctx context.Context, docID string) error {
	doc, err := i.db.GetDocumentByID(ctx, docID)
	if err != nil || doc == nil {
		return fmt.Errorf("document not found: %w", err)
	}

	_ = i.db.UpdateDocumentStatus(ctx, docID, "processing")

	bucket, key := parseS3URL(doc.StorageURL)

	// get streaming reader from object storage
	rc, err := i.obj.GetObjectReader(ctx, bucket, key)
	if err != nil {
		_ = i.db.UpdateDocumentStatus(ctx, docID, "failed")
		return fmt.Errorf("get object reader: %w", err)
	}
	defer rc.Close()

	// Build an errgroup to tie the pipeline stages together.
	g, gctx := errgroup.WithContext(ctx)

	// extract ->  fragments (receive-only channel).
	fragCh := i.streamExtract(gctx, g, rc, i.cfg.MaxFragmentLen)

	// fragments -> chunks (receive-only channel).
	chunkCh := i.streamChunk(gctx, g, fragCh, i.cfg.TargetTokens, i.cfg.OverlapTokens)

	// chunks → embed + persist.
	g.Go(func() error {
		return i.embedAndPersist(gctx, docID, chunkCh, i.cfg.BatchSize, i.cfg.EmbedDim)
	})

	// Wait for all stages. Any error cancels the rest.
	if err := g.Wait(); err != nil {
		_ = i.db.UpdateDocumentStatus(ctx, docID, "failed")
		return err
	}

	// Success.
	return i.db.UpdateDocumentStatus(ctx, docID, "ready")
}

// parseS3URL extracts the bucket and key from a typical virtual-hosted–style S3 URL.
// Example: https://my-bucket.s3.us-east-2.amazonaws.com/path/to/file.pdf
func parseS3URL(u string) (bucket, key string) {
	// Very small parser; if you already store bucket/key separately, prefer those fields.
	// You can replace this with a robust parser later.
	hostPath := strings.SplitN(strings.TrimPrefix(u, "https://"), "/", 2)
	host := hostPath[0]
	if len(hostPath) == 2 {
		key = hostPath[1]
	}
	parts := strings.Split(host, ".")
	if len(parts) > 0 {
		bucket = parts[0]
	}
	return bucket, key
}
