package nodes

import (
	"bufio"
	"bytes"
	"context"
	"net/http"
	"strconv"
	"strings"

	"christiangeorgelucas/http-message-tools/axiom"
	gen "christiangeorgelucas/http-message-tools/gen"
)

// Parse a raw HTTP/1.x response message (status-line, headers, body) into
// structured form. Delegates the status-line grammar plus Content-Length /
// chunked body decoding to net/http.ReadResponse, and separately re-scans
// the header block to preserve wire order and repeated headers (e.g.
// multiple Set-Cookie headers), which net/http's map-based Header type
// does not retain. Set request_method="HEAD" when this response answers a
// HEAD request, since a HEAD response's Content-Length describes a body
// that is not actually present — RFC 7230 leaves that ambiguous without
// knowing the request method. Rejects a message with both Content-Length
// and Transfer-Encoding present and bounds total input to 4 MiB.
func ParseResponse(ctx context.Context, ax axiom.Context, input *gen.ParseResponseInput) (*gen.ParseResponseOutput, error) {
	if input == nil || len(input.Data) == 0 {
		return &gen.ParseResponseOutput{Error: "data is required"}, nil
	}
	if len(input.Data) > maxInputBytes {
		return &gen.ParseResponseOutput{Error: "input exceeds the 4 MiB size limit"}, nil
	}

	_, rest, ok := splitStartLine(input.Data)
	if !ok {
		return &gen.ParseResponseOutput{Error: "no line terminator found; not a valid HTTP message"}, nil
	}
	hs, herr := scanOrderedHeaders(rest)
	if herr != nil {
		return &gen.ParseResponseOutput{Error: herr.Error()}, nil
	}
	if hasAmbiguousFraming(hs) {
		return &gen.ParseResponseOutput{Error: "ambiguous message framing: both Content-Length and Transfer-Encoding present"}, nil
	}

	var req *http.Request
	if input.RequestMethod != "" {
		req = &http.Request{Method: strings.ToUpper(input.RequestMethod)}
	}
	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(input.Data)), req)
	if err != nil {
		return &gen.ParseResponseOutput{Error: "parse error: " + err.Error()}, nil
	}

	body, truncated, berr := readBodyBounded(resp.Body, clampMaxBody(input.MaxBodyBytes))
	if berr != nil {
		return &gen.ParseResponseOutput{Error: "body read error: " + berr.Error()}, nil
	}

	return &gen.ParseResponseOutput{
		Ok:            true,
		Version:       resp.Proto,
		StatusCode:    int32(resp.StatusCode),
		Reason:        strings.TrimSpace(strings.TrimPrefix(resp.Status, strconv.Itoa(resp.StatusCode))),
		Headers:       headersToProto(hs),
		Body:          body,
		BodyTruncated: truncated,
	}, nil
}
