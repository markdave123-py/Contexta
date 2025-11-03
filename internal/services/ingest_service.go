package services

// import (
// 	"context"
// 	"fmt"
// 	"strings"

// 	"golang.org/x/sync/errgroup"

// 	"github.com/markdave123-py/Contexta/internal/core"
// 	"github.com/markdave123-py/Contexta/internal/models"
// )

// // chunk is the internal representation passed through the pipeline.
// //
// // Pos:      stable, zero-based position of the chunk inside the document.
// // Text:     chunk content (built from one or more fragments).
// // TokenCnt: approximate token count (used for batching and overlap math).
// type chunk struct {
// 	Pos      int
// 	Text     string
// 	TokenCnt int
// }

// // DocumentIngestor orchestrates the background ingestion pipeline:
// //
// // db:        persistence for document and chunks.
// // obj:       object storage for streaming large files.
// // embedder:  embedding provider (Gemini/OpenAI/etc).
// // cfg:       runtime tuning knobs for the pipeline.
// // jobs:      in-memory queue of document IDs to process (easy to swap with Kafka later).
// type DocumentIngestor struct {
// 	db       core.DbClient
// 	obj      core.ObjectClient
// 	embedder core.EmbeddingProvider
// 	cfg      IngestConfig
// 	jobs     chan string
// }

// // NewDocumentIngestor constructs the ingestor with a bounded job queue (64).
// func NewDocumentIngestor(db core.DbClient, obj core.ObjectClient, emb core.EmbeddingProvider, cfg IngestConfig) *DocumentIngestor {
// 	return &DocumentIngestor{
// 		db: db, obj: obj, embedder: emb, cfg: cfg,
// 		jobs: make(chan string, 64),
// 	}
// }

// // Start runs a single worker goroutine reading from the jobs channel.
// func (i *DocumentIngestor) Start(ctx context.Context) {
// 	go func() {
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				return
// 			case docID := <-i.jobs:
// 				// Best-effort processing; errors can be logged or reported.
// 				_ = i.processOne(ctx, docID)
// 			}
// 		}
// 	}()
// }

// // Enqueue schedules a document ID for ingestion.
// // If the queue is full, this call will block until space frees up.
// func (i *DocumentIngestor) Enqueue(docID string) {
// 	i.jobs <- docID
// }

// // processOne streams, chunks, embeds and persists for a single document ID.
// func (i *DocumentIngestor) processOne(ctx context.Context, docID string) error {
// 	doc, err := i.db.GetDocumentByID(ctx, docID)
// 	if err != nil || doc == nil {
// 		return fmt.Errorf("document not found: %w", err)
// 	}

// 	_ = i.db.UpdateDocumentStatus(ctx, docID, "processing")

// 	bucket, key := parseS3URL(doc.StorageURL)

// 	// get streaming reader from object storage
// 	rc, err := i.obj.GetObjectReader(ctx, bucket, key)
// 	if err != nil {
// 		_ = i.db.UpdateDocumentStatus(ctx, docID, "failed")
// 		return fmt.Errorf("get object reader: %w", err)
// 	}
// 	defer rc.Close()

// 	// Build an errgroup to tie the pipeline stages together.
// 	g, gctx := errgroup.WithContext(ctx)

// 	// extract ->  fragments (receive-only channel).
// 	fragCh := i.streamExtract(gctx, g, rc, i.cfg.MaxFragmentLen)

// 	// fragments -> chunks (receive-only channel).
// 	chunkCh := i.streamChunk(gctx, g, fragCh, i.cfg.TargetTokens, i.cfg.OverlapTokens)

// 	// chunks → embed + persist.
// 	g.Go(func() error {
// 		return i.embedAndPersist(gctx, docID, chunkCh, i.cfg.BatchSize, i.cfg.EmbedDim)
// 	})

// 	// Wait for all stages. Any error cancels the rest.
// 	if err := g.Wait(); err != nil {
// 		_ = i.db.UpdateDocumentStatus(ctx, docID, "failed")
// 		return err
// 	}

// 	// Success.
// 	return i.db.UpdateDocumentStatus(ctx, docID, "ready")
// }

// // streamExtract converts an io.Reader into a stream of small text fragments.
// //
// // r:           the streaming source (e.g., S3 GetObject Body).
// // maxFragLen:  soft cap to avoid very large messages; long lines split into multiple fragments.
// // out:         receive-only channel of fragments; closed when extraction completes.
// // errgroup:    manages lifecycle; any error cancels the whole pipeline.
// // func (i *DocumentIngestor) streamExtract(
// // 	ctx context.Context,
// // 	g *errgroup.Group,
// // 	r io.Reader,
// // 	maxFragLen int,
// // ) <-chan string {
// // 	out := make(chan string, 8)

// // 	g.Go(func() error {
// // 		defer close(out)

// // 		sc := bufio.NewScanner(r)

// // 		// Raise the maximum token size Scanner will handle (default is 64K).
// // 		// We allow up to 1MB lines to be safe with texty PDFs; adjust if needed.
// // 		buf := make([]byte, 0, 64*1024)
// // 		sc.Buffer(buf, 1<<20)

// // 		for sc.Scan() {
// // 			// Cancellation check so we stop early if downstream failed.
// // 			select {
// // 			case <-ctx.Done():
// // 				return ctx.Err()
// // 			default:
// // 			}

// // 			line := sc.Text()
// // 			if line == "" {
// // 				continue
// // 			}

// // 			// Split the line into bounded fragments to avoid huge messages.
// // 			for len(line) > maxFragLen {
// // 				frag := line[:maxFragLen]
// // 				line = line[maxFragLen:]

// // 				select {
// // 				case out <- frag:
// // 				case <-ctx.Done():
// // 					return ctx.Err()
// // 				}
// // 			}
// // 			// Emit the tail.
// // 			select {
// // 			case out <- line:
// // 			case <-ctx.Done():
// // 				return ctx.Err()
// // 			}
// // 		}
// // 		// Propagate scanner errors (I/O errors from the reader, etc.)
// // 		return sc.Err()
// // 	})

// // 	return out
// // }

// // streamChunk groups incoming fragments into token-bounded chunks with optional overlap.
// //
// // frags:          upstream fragments channel.
// // targetTokens:   approximate tokens per chunk.
// // overlapTokens:  tokens to retain from the end of the previous chunk as seed of the next (e.g., 50).
// // out:            receive-only channel of chunk structs with Pos/Text/TokenCnt.
// // func (i *DocumentIngestor) streamChunk(
// // 	ctx context.Context,
// // 	g *errgroup.Group,
// // 	frags <-chan string,
// // 	targetTokens int,
// // 	overlapTokens int,
// // ) <-chan chunk {
// // 	out := make(chan chunk, 8)

// // 	g.Go(func() error {
// // 		defer close(out)

// // 		var buf []string // rolling fragment buffer
// // 		var tokSum int   // estimated tokens in the buffer
// // 		pos := 0         // emitted chunk position

// // 		// flush emits the current buffer as a chunk and prepares the buffer for the next chunk,
// // 		// preserving overlapTokens from the tail if configured.
// // 		flush := func() error {
// // 			if tokSum == 0 {
// // 				return nil
// // 			}
// // 			text := strings.Join(buf, "\n")
// // 			ch := chunk{Pos: pos, Text: text, TokenCnt: tokSum}
// // 			pos++

// // 			// Emit the chunk to downstream; backpressure applies here.
// // 			select {
// // 			case out <- ch:
// // 			case <-ctx.Done():
// // 				return ctx.Err()
// // 			}

// // 			// Compute overlap: keep a tail whose token sum ≈ overlapTokens.
// // 			if overlapTokens > 0 {
// // 				keep := []string{}
// // 				remain := overlapTokens
// // 				for j := len(buf) - 1; j >= 0 && remain > 0; j-- {
// // 					t := approxTokens(buf[j])
// // 					keep = append([]string{buf[j]}, keep...) // prepend to keep original order
// // 					remain -= t
// // 				}
// // 				buf = keep

// // 				// Recompute tokSum for the kept tail.
// // 				tokSum = 0
// // 				for _, s := range buf {
// // 					tokSum += approxTokens(s)
// // 				}
// // 			} else {
// // 				// No overlap: clear buffer.
// // 				buf = buf[:0]
// // 				tokSum = 0
// // 			}
// // 			return nil
// // 		}

// // 		for frag := range frags {
// // 			// Cancel early if downstream failed.
// // 			select {
// // 			case <-ctx.Done():
// // 				return ctx.Err()
// // 			default:
// // 			}

// // 			// Accumulate fragment and its token estimate.
// // 			t := approxTokens(frag)
// // 			buf = append(buf, frag)
// // 			tokSum += t

// // 			// If we've reached the target, emit a chunk.
// // 			if tokSum >= targetTokens {
// // 				if err := flush(); err != nil {
// // 					return err
// // 				}
// // 			}
// // 		}

// // 		// Emit remaining tail (if any).
// // 		if err := flush(); err != nil {
// // 			return err
// // 		}
// // 		return nil
// // 	})

// // 	return out
// // }

// // embedAndPersist consumes chunks, embeds them in batches, and writes to DB.
// // This function provides the downstream sink for the pipeline above.
// //
// // docID:      current document ID.
// // in:         chunk stream from streamChunk.
// // batchSize:  number of chunks to embed/write per batch (limits memory).
// // embedDim:   model dimension (0 = default).
// func (i *DocumentIngestor) embedAndPersist(
// 	ctx context.Context,
// 	docID string,
// 	in <-chan chunk,
// 	batchSize int,
// 	embedDim int,
// ) error {
// 	batch := make([]chunk, 0, batchSize)

// 	// flush embeds the current batch and inserts it into the database.
// 	flush := func(items []chunk) error {
// 		if len(items) == 0 {
// 			return nil
// 		}

// 		texts := make([]string, len(items))
// 		for idx := range items {
// 			texts[idx] = items[idx].Text
// 		}

// 		vecs, err := i.embedder.EmbedTexts(ctx, texts, embedDim)
// 		if err != nil {
// 			return fmt.Errorf("embed: %w", err)
// 		}
// 		if len(vecs) != len(items) {
// 			return fmt.Errorf("embed size mismatch: got %d want %d", len(vecs), len(items))
// 		}

// 		// 3) Map to persistence rows and write once.
// 		rows := make([]models.DocumentChunk, len(items))
// 		for k := range items {
// 			rows[k] = models.DocumentChunk{
// 				DocumentID: docID,
// 				Text:       items[k].Text,
// 				Embedding:  vecs[k],
// 				Position:   items[k].Pos,
// 				TokenCount: items[k].TokenCnt,
// 			}
// 		}
// 		if err := i.db.InsertDocumentChunks(ctx, rows); err != nil {
// 			return fmt.Errorf("insert chunks: %w", err)
// 		}
// 		return nil
// 	}

// 	// Read the stream and flush in batches.
// 	for c := range in {
// 		batch = append(batch, c)
// 		if len(batch) == batchSize {
// 			if err := flush(batch); err != nil {
// 				return err
// 			}
// 			batch = batch[:0]
// 		}
// 	}
// 	// Final tail.
// 	if err := flush(batch); err != nil {
// 		return err
// 	}
// 	return nil
// }

// // parseS3URL extracts the bucket and key from a typical virtual-hosted–style S3 URL.
// // Example: https://my-bucket.s3.us-east-2.amazonaws.com/path/to/file.pdf
// func parseS3URL(u string) (bucket, key string) {
// 	// Very small parser; if you already store bucket/key separately, prefer those fields.
// 	// You can replace this with a robust parser later.
// 	hostPath := strings.SplitN(strings.TrimPrefix(u, "https://"), "/", 2)
// 	host := hostPath[0]
// 	if len(hostPath) == 2 {
// 		key = hostPath[1]
// 	}
// 	parts := strings.Split(host, ".")
// 	if len(parts) > 0 {
// 		bucket = parts[0]
// 	}
// 	return bucket, key
// }
