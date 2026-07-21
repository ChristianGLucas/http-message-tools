package nodes

import (
	"context"
	"strings"

	"christiangeorgelucas/http-message-tools/axiom"
	gen "christiangeorgelucas/http-message-tools/gen"
)

// Split a request-target (e.g. "/search?q=go+http") into its path and raw
// query string. A minimal split on the first '?' — the query string itself
// is returned undecoded; hand it to url-tools if you need it parsed into
// key/value pairs.
func SplitRequestTarget(ctx context.Context, ax axiom.Context, input *gen.SplitRequestTargetInput) (*gen.SplitRequestTargetOutput, error) {
	if input == nil || input.Target == "" {
		return &gen.SplitRequestTargetOutput{Error: "target is required"}, nil
	}
	if len(input.Target) > maxLineBytes {
		return &gen.SplitRequestTargetOutput{Error: "target exceeds the 8192 byte limit"}, nil
	}

	path, query, _ := strings.Cut(input.Target, "?")
	return &gen.SplitRequestTargetOutput{Ok: true, Path: path, Query: query}, nil
}
