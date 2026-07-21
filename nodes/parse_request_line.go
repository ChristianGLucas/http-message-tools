package nodes

import (
	"bufio"
	"context"
	"net/http"
	"strings"

	"christiangeorgelucas/http-message-tools/axiom"
	gen "christiangeorgelucas/http-message-tools/gen"
)

// Parse a bare request-line, e.g. "GET /path?q=1 HTTP/1.1", into its method,
// request-target, and HTTP version. Reuses net/http.ReadRequest for the
// actual grammar by synthesizing a minimal complete request (the line plus
// an empty header block) rather than hand-splitting on spaces, so the same
// validation ParseRequest applies here too.
func ParseRequestLine(ctx context.Context, ax axiom.Context, input *gen.ParseRequestLineInput) (*gen.ParseRequestLineOutput, error) {
	if input == nil || input.Line == "" {
		return &gen.ParseRequestLineOutput{Error: "line is required"}, nil
	}
	if len(input.Line) > maxLineBytes {
		return &gen.ParseRequestLineOutput{Error: "line exceeds the 8192 byte limit"}, nil
	}
	if strings.ContainsAny(input.Line, "\r\n") {
		return &gen.ParseRequestLineOutput{Error: "line must not contain embedded newlines"}, nil
	}

	synthetic := input.Line + "\r\n\r\n"
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(synthetic)))
	if err != nil {
		return &gen.ParseRequestLineOutput{Error: "parse error: " + err.Error()}, nil
	}

	return &gen.ParseRequestLineOutput{
		Ok:      true,
		Method:  req.Method,
		Target:  req.RequestURI,
		Version: req.Proto,
	}, nil
}
