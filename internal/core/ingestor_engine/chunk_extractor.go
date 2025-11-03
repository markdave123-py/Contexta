package ingestorengine

import (
	"context"
	"strings"

	"golang.org/x/sync/errgroup"
)

// streamChunk groups incoming fragments into token-bounded chunks with optional overlap.
//
// frags:          upstream fragments channel.
// targetTokens:   approximate tokens per chunk.
// overlapTokens:  tokens to retain from the end of the previous chunk as seed of the next (e.g., 50).
// out:            receive-only channel of chunk structs with Pos/Text/TokenCnt.
func (i *DocumentIngestor) streamChunk(
	ctx context.Context,
	g *errgroup.Group,
	frags <-chan string,
	targetTokens int,
	overlapTokens int,
) <-chan chunk {
	out := make(chan chunk, 8)

	g.Go(func() error {
		defer close(out)

		var buf []string // rolling fragment buffer
		var tokSum int   // estimated tokens in the buffer
		pos := 0         // emitted chunk position

		// flush emits the current buffer as a chunk and prepares the buffer for the next chunk,
		// preserving overlapTokens from the tail if configured.
		flush := func() error {
			if tokSum == 0 {
				return nil
			}
			text := strings.Join(buf, "\n")
			ch := chunk{Pos: pos, Text: text, TokenCnt: tokSum}
			pos++

			// Emit the chunk to downstream; backpressure applies here.
			select {
			case out <- ch:
			case <-ctx.Done():
				return ctx.Err()
			}

			// Compute overlap: keep a tail whose token sum ≈ overlapTokens.
			if overlapTokens > 0 {
				keep := []string{}
				remain := overlapTokens
				for j := len(buf) - 1; j >= 0 && remain > 0; j-- {
					t := approxTokens(buf[j])
					keep = append([]string{buf[j]}, keep...) // prepend to keep original order
					remain -= t
				}
				buf = keep

				// Recompute tokSum for the kept tail.
				tokSum = 0
				for _, s := range buf {
					tokSum += approxTokens(s)
				}
			} else {
				// No overlap: clear buffer.
				buf = buf[:0]
				tokSum = 0
			}
			return nil
		}

		for frag := range frags {
			// Cancel early if downstream failed.
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// Accumulate fragment and its token estimate.
			t := approxTokens(frag)
			buf = append(buf, frag)
			tokSum += t

			// If we've reached the target, emit a chunk.
			if tokSum >= targetTokens {
				if err := flush(); err != nil {
					return err
				}
			}
		}

		// Emit remaining tail (if any).
		if err := flush(); err != nil {
			return err
		}
		return nil
	})

	return out
}


// approxTokens is a cheap token estimator (~4 chars ≈ 1 token).
// Replace with a real tokenizer later to improve chunk boundaries.
func approxTokens(s string) int {
	n := len([]rune(s))
	if n <= 0 {
		return 0
	}
	return (n + 3) / 4
}