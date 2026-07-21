package nodes

import (
	"bufio"
	"bytes"
	"context"
	"net/http"

	"christiangeorgelucas/http-message-tools/axiom"
	gen "christiangeorgelucas/http-message-tools/gen"
)

// Build a raw HTTP/1.x request message from structured components — the
// inverse of ParseRequest. Every header name/value, the method, and the
// target are validated against RFC 7230's token/field-content grammar
// before assembly, rejecting anything containing a CR or LF (which would
// otherwise let a caller inject extra headers or a second message into the
// output — classic CRLF/header injection). The assembled bytes are then
// re-parsed with net/http.ReadRequest before being returned, so this node
// never emits a message its own ParseRequest would reject.
func BuildRequest(ctx context.Context, ax axiom.Context, input *gen.BuildRequestInput) (*gen.BuildRequestOutput, error) {
	if input == nil || input.Method == "" {
		return &gen.BuildRequestOutput{Error: "method is required"}, nil
	}
	if !isValidToken(input.Method) {
		return &gen.BuildRequestOutput{Error: "invalid method: must be an RFC 7230 token"}, nil
	}
	target := input.Target
	if target == "" {
		target = "/"
	}
	if !isValidTargetOrVersion(target) {
		return &gen.BuildRequestOutput{Error: "invalid target: must not contain whitespace or control characters"}, nil
	}
	version := normalizeVersion(input.Version)
	if !validVersion(version) {
		return &gen.BuildRequestOutput{Error: "invalid version: expected a form like \"HTTP/1.1\""}, nil
	}
	if err := validateHeadersForBuild(input.Headers); err != nil {
		return &gen.BuildRequestOutput{Error: err.Error()}, nil
	}
	if len(input.Body) > maxInputBytes {
		return &gen.BuildRequestOutput{Error: "body exceeds the 4 MiB size limit"}, nil
	}

	var buf bytes.Buffer
	buf.WriteString(input.Method)
	buf.WriteByte(' ')
	buf.WriteString(target)
	buf.WriteByte(' ')
	buf.WriteString(version)
	buf.WriteString("\r\n")
	writeHeaders(&buf, input.Headers)
	buf.WriteString("\r\n")
	buf.Write(input.Body)

	data := buf.Bytes()
	if len(data) > maxInputBytes {
		return &gen.BuildRequestOutput{Error: "assembled message exceeds the 4 MiB size limit"}, nil
	}

	// Self-validate: never hand back bytes our own parser would reject.
	// Reading the body (not just the start-line/headers) matters here: a
	// caller-supplied Content-Length that does not match the actual body
	// length (or a Transfer-Encoding: chunked header on a body that isn't
	// actually chunk-encoded) would otherwise slip through undetected,
	// since http.ReadRequest returns before its lazily-read Body is
	// consumed.
	reReq, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(data)))
	if err != nil {
		return &gen.BuildRequestOutput{Error: "assembled message failed to re-parse: " + err.Error()}, nil
	}
	if _, _, err := readBodyBounded(reReq.Body, maxInputBytes); err != nil {
		return &gen.BuildRequestOutput{Error: "assembled message's body failed to re-parse (likely a Content-Length/Transfer-Encoding mismatch): " + err.Error()}, nil
	}

	return &gen.BuildRequestOutput{Ok: true, Data: data}, nil
}
