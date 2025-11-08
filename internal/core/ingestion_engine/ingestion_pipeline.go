package ingestion_engine

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/markdave123-py/Contexta/internal/core"
	"golang.org/x/sync/errgroup"
)

// NewDocumentIngestor constructs the ingestor with a bounded job queue (64).
func NewDocumentIngestor(db core.DbClient, obj core.ObjectClient, emb core.EmbeddingProvider, extrator core.DocumentExtractor, cfg *IngestConfig) *DocumentIngestor {
	return &DocumentIngestor{
		db: db, obj: obj, embedder: emb, cfg: cfg, extrator: extrator,
		jobs: make(chan string, 64),
	}
}

// Start runs a single worker goroutine reading from the jobs channel.
// It ochestrate the pipeline that extract, parse, embed and persist docs.
func (i *DocumentIngestor) Start(ctx context.Context, numWorkers int) {

	for w := 1; w <= numWorkers; w++ {
		go func(w int) {
			for {
				select {
				case <-ctx.Done():
					log.Println("DocumentIngestor: Worker shutting down.")
					return
				case docID := <-i.jobs:
					log.Printf("DocumentIngestor: Processing document %s by worker with ID %d", docID, w)

					if err := i.processOne(ctx, docID); err != nil {
						log.Printf("DocumentIngestor: Error processing document %s: %v", docID, err)
					}
				}
			}
		}(w)
	}
}

// Enqueue schedules a document ID for ingestion.
// If the queue is full, this call will block until space frees up.
func (i *DocumentIngestor) Enqueue(docID string) {
	i.jobs <- docID
}

// processOne streams, chunks, embeds and persists for a single document ID.
func (i *DocumentIngestor) processOne(ctx context.Context, docID string) error { // Create a fresh context with longer timeout for processing
	proctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	doc, err := i.db.GetDocumentByID(ctx, docID)
	if err != nil || doc == nil {
		return fmt.Errorf("document not found: %w", err)
	}

	bucket, key := parseS3URL(doc.StorageURL)

	// get streaming reader from object storage
	rc, err := i.obj.GetFile(proctx, bucket, key)
	if err != nil {
		_ = i.db.UpdateDocumentStatus(ctx, docID, "failed")
		return fmt.Errorf("get object reader: %w", err)
	}

	// Build an errgroup to tie the pipeline stages together.
	g, gctx := errgroup.WithContext(context.Background())

	// extract documents ->  fragments (receive-only channel).
	fragCh, err := i.extrator.ExtractText(gctx, g, rc, doc.ContentType)

	// fragments -> chunks (receive-only channel).
	chunkCh := i.streamChunk(gctx, g, fragCh, i.cfg.TargetTokens, i.cfg.OverlapTokens)

	// chunks → embed + persist.
	g.Go(func() error {
		return i.embedAndPersist(gctx, docID, chunkCh, i.cfg.BatchSize)
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
