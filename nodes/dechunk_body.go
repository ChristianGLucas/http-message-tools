package nodes

import (
	"bytes"
	"context"
	"net/http/httputil"

	"christiangeorgelucas/http-message-tools/axiom"
	gen "christiangeorgelucas/http-message-tools/gen"
)

// Decode a Transfer-Encoding: chunked message body into its dechunked
// bytes, via net/http/httputil.NewChunkedReader — the same decoder
// net/http itself uses to transparently dechunk a response body. Output is
// bounded to max_output_bytes (default/max 4 MiB): a chunk-size line can
// declare an arbitrarily large chunk, but decoding is read incrementally
// and stops the moment the cap is reached, so a hostile declared size
// never drives an allocation anywhere near that size.
func DechunkBody(ctx context.Context, ax axiom.Context, input *gen.DechunkBodyInput) (*gen.DechunkBodyOutput, error) {
	if input == nil || len(input.Data) == 0 {
		return &gen.DechunkBodyOutput{Error: "data is required"}, nil
	}
	if len(input.Data) > maxInputBytes {
		return &gen.DechunkBodyOutput{Error: "input exceeds the 4 MiB size limit"}, nil
	}

	cr := httputil.NewChunkedReader(bytes.NewReader(input.Data))
	out, truncated, err := readBodyBounded(cr, clampMaxOutput(input.MaxOutputBytes))
	if err != nil {
		return &gen.DechunkBodyOutput{Error: "chunked decode error: " + err.Error()}, nil
	}

	return &gen.DechunkBodyOutput{Ok: true, Data: out, Truncated: truncated}, nil
}
