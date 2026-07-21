package nodes

import (
	"context"

	"christiangeorgelucas/http-message-tools/axiom"
	gen "christiangeorgelucas/http-message-tools/gen"
)

// Parse a raw header block (zero or more "Name: value" lines, no leading
// request-line/status-line) into a structured, ORDER-preserving list —
// repeated headers of the same name appear as separate entries in wire
// order, and obsolete line-folding (a continuation line starting with a
// space or tab) is honored via net/textproto's continued-line reader.
func ParseHeaderBlock(ctx context.Context, ax axiom.Context, input *gen.ParseHeaderBlockInput) (*gen.ParseHeaderBlockOutput, error) {
	if input == nil {
		return &gen.ParseHeaderBlockOutput{Error: "data is required"}, nil
	}
	if len(input.Data) > maxInputBytes {
		return &gen.ParseHeaderBlockOutput{Error: "input exceeds the 4 MiB size limit"}, nil
	}

	hs, err := scanOrderedHeaders(input.Data)
	if err != nil {
		return &gen.ParseHeaderBlockOutput{Error: err.Error()}, nil
	}

	return &gen.ParseHeaderBlockOutput{
		Ok:      true,
		Headers: headersToProto(hs),
	}, nil
}
