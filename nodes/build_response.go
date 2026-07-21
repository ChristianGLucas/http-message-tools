package nodes

import (
	"bufio"
	"bytes"
	"context"
	"net/http"
	"strconv"

	"christiangeorgelucas/http-message-tools/axiom"
	gen "christiangeorgelucas/http-message-tools/gen"
)

// Build a raw HTTP/1.x response message from structured components — the
// inverse of ParseResponse. reason defaults to the standard RFC 9110
// reason phrase for status_code (via net/http.StatusText) when empty; an
// unrecognized code with no table entry and no explicit reason is
// rejected. Every header name/value is validated against RFC 7230's
// token/field-content grammar before assembly, and the assembled bytes are
// re-parsed with net/http.ReadResponse before being returned, so this node
// never emits a message its own ParseResponse would reject.
func BuildResponse(ctx context.Context, ax axiom.Context, input *gen.BuildResponseInput) (*gen.BuildResponseOutput, error) {
	if input == nil || input.StatusCode == 0 {
		return &gen.BuildResponseOutput{Error: "status_code is required"}, nil
	}
	if input.StatusCode < 100 || input.StatusCode > 599 {
		return &gen.BuildResponseOutput{Error: "status_code must be between 100 and 599"}, nil
	}
	reason := input.Reason
	if reason == "" {
		reason = statusText(int(input.StatusCode))
		if reason == "" {
			return &gen.BuildResponseOutput{Error: "status_code has no standard reason phrase; supply reason explicitly"}, nil
		}
	}
	if !isValidFieldValue(reason) {
		return &gen.BuildResponseOutput{Error: "invalid reason: must not contain control characters"}, nil
	}
	version := normalizeVersion(input.Version)
	if !validVersion(version) {
		return &gen.BuildResponseOutput{Error: "invalid version: expected a form like \"HTTP/1.1\""}, nil
	}
	if err := validateHeadersForBuild(input.Headers); err != nil {
		return &gen.BuildResponseOutput{Error: err.Error()}, nil
	}
	if len(input.Body) > maxInputBytes {
		return &gen.BuildResponseOutput{Error: "body exceeds the 4 MiB size limit"}, nil
	}

	var buf bytes.Buffer
	buf.WriteString(version)
	buf.WriteByte(' ')
	buf.WriteString(strconv.Itoa(int(input.StatusCode)))
	buf.WriteByte(' ')
	buf.WriteString(reason)
	buf.WriteString("\r\n")
	writeHeaders(&buf, input.Headers)
	buf.WriteString("\r\n")
	buf.Write(input.Body)

	data := buf.Bytes()
	if len(data) > maxInputBytes {
		return &gen.BuildResponseOutput{Error: "assembled message exceeds the 4 MiB size limit"}, nil
	}

	// Self-validate: never hand back bytes our own parser would reject.
	// Reading the body (not just the start-line/headers) matters here: a
	// caller-supplied Content-Length that does not match the actual body
	// length (or a Transfer-Encoding: chunked header on a body that isn't
	// actually chunk-encoded) would otherwise slip through undetected,
	// since http.ReadResponse returns before its lazily-read Body is
	// consumed.
	reResp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(data)), nil)
	if err != nil {
		return &gen.BuildResponseOutput{Error: "assembled message failed to re-parse: " + err.Error()}, nil
	}
	if _, _, err := readBodyBounded(reResp.Body, maxInputBytes); err != nil {
		return &gen.BuildResponseOutput{Error: "assembled message's body failed to re-parse (likely a Content-Length/Transfer-Encoding mismatch): " + err.Error()}, nil
	}

	return &gen.BuildResponseOutput{Ok: true, Data: data}, nil
}
