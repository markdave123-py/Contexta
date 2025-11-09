package ingestion_engine

import "context"

type Ingestor interface {
	Start(ctx context.Context, numWorkers int)
	Enqueue(docID string)
	ProcessOne(ctx context.Context, docID string) error
}
