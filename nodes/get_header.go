package nodes

import (
	"context"
	"strings"

	"christiangeorgelucas/http-message-tools/axiom"
	gen "christiangeorgelucas/http-message-tools/gen"
)

// Extract a single named header's value(s), case-insensitively, from
// either a bare header block or a complete HTTP message (a leading
// request-line/status-line, if present, is detected and skipped). Returns
// every occurrence in wire order — useful for a repeated header like
// Set-Cookie — plus first_value as a plain-scalar convenience for
// composing into a downstream node's scalar input across a flow edge.
func GetHeader(ctx context.Context, ax axiom.Context, input *gen.GetHeaderInput) (*gen.GetHeaderOutput, error) {
	if input == nil || input.Name == "" {
		return &gen.GetHeaderOutput{Error: "name is required"}, nil
	}
	if len(input.Data) > maxInputBytes {
		return &gen.GetHeaderOutput{Error: "input exceeds the 4 MiB size limit"}, nil
	}

	hs, err := scanHeadersFlexible(input.Data)
	if err != nil {
		return &gen.GetHeaderOutput{Error: err.Error()}, nil
	}

	var values []string
	for _, h := range hs {
		if strings.EqualFold(h.name, input.Name) {
			values = append(values, h.value)
		}
	}

	out := &gen.GetHeaderOutput{Ok: true, Found: len(values) > 0, Values: values}
	if len(values) > 0 {
		out.FirstValue = values[0]
	}
	return out, nil
}
