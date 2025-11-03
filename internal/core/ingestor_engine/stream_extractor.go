package ingestorengine

import (
	"bufio"
	"context"
	"io"

	"golang.org/x/sync/errgroup"
)

// streamExtract converts an io.Reader into a stream of small text fragments.
//
// r:           the streaming source (e.g., S3 GetObject Body).
// maxFragLen:  soft cap to avoid very large messages; long lines split into multiple fragments.
// out:         receive-only channel of fragments; closed when extraction completes.
// errgroup:    manages lifecycle; any error cancels the whole pipeline.
func (i *DocumentIngestor) streamExtract(
	ctx context.Context,
	g *errgroup.Group,
	r io.Reader,
	maxFragLen int,
) <-chan string {
	out := make(chan string, 8)

	g.Go(func() error {
		defer close(out)

		sc := bufio.NewScanner(r)

		// Raise the maximum token size Scanner will handle (default is 64K).
		// We allow up to 1MB lines to be safe with texty PDFs; adjust if needed.
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 1<<20)

		for sc.Scan() {
			// Cancellation check so we stop early if downstream failed.
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			line := sc.Text()
			if line == "" {
				continue
			}

			// Split the line into bounded fragments to avoid huge messages.
			for len(line) > maxFragLen {
				frag := line[:maxFragLen]
				line = line[maxFragLen:]

				select {
				case out <- frag:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			// Emit the tail.
			select {
			case out <- line:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		// Propagate scanner errors (I/O errors from the reader, etc.)
		return sc.Err()
	})

	return out
}
