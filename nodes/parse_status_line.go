package nodes

import (
	"bufio"
	"context"
	"net/http"
	"strconv"
	"strings"

	"christiangeorgelucas/http-message-tools/axiom"
	gen "christiangeorgelucas/http-message-tools/gen"
)

// Parse a bare status-line, e.g. "HTTP/1.1 404 Not Found", into its HTTP
// version, status code, and reason phrase. Reuses net/http.ReadResponse for
// the actual grammar by synthesizing a minimal complete response (the line
// plus an empty header block) rather than hand-splitting on spaces.
func ParseStatusLine(ctx context.Context, ax axiom.Context, input *gen.ParseStatusLineInput) (*gen.ParseStatusLineOutput, error) {
	if input == nil || input.Line == "" {
		return &gen.ParseStatusLineOutput{Error: "line is required"}, nil
	}
	if len(input.Line) > maxLineBytes {
		return &gen.ParseStatusLineOutput{Error: "line exceeds the 8192 byte limit"}, nil
	}
	if strings.ContainsAny(input.Line, "\r\n") {
		return &gen.ParseStatusLineOutput{Error: "line must not contain embedded newlines"}, nil
	}

	synthetic := input.Line + "\r\n\r\n"
	resp, err := http.ReadResponse(bufio.NewReader(strings.NewReader(synthetic)), nil)
	if err != nil {
		return &gen.ParseStatusLineOutput{Error: "parse error: " + err.Error()}, nil
	}

	return &gen.ParseStatusLineOutput{
		Ok:         true,
		Version:    resp.Proto,
		StatusCode: int32(resp.StatusCode),
		Reason:     strings.TrimSpace(strings.TrimPrefix(resp.Status, strconv.Itoa(resp.StatusCode))),
	}, nil
}
