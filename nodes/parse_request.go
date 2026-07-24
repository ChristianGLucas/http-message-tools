package nodes

import (
	"bufio"
	"bytes"
	"context"
	"net/http"

	"christiangeorgelucas/http-message-tools/axiom"
	gen "christiangeorgelucas/http-message-tools/gen"
)

// Parse a raw HTTP/1.x request message (request-line, headers, body) into
// structured form. Delegates the request-line grammar plus Content-Length /
// chunked body decoding to net/http.ReadRequest, and separately re-scans the
// header block to preserve wire ORDER and repeated headers, which net/http's
// map-based Header type does not retain. Rejects a message with both
// Content-Length and Transfer-Encoding present (a request-smuggling
// ambiguity net/http silently resolves by dropping both headers rather than
// flagging).
func ParseRequest(ctx context.Context, ax axiom.Context, input *gen.ParseRequestInput) (*gen.ParseRequestOutput, error) {
	if input == nil || len(input.Data) == 0 {
		return &gen.ParseRequestOutput{Error: "data is required"}, nil
	}
	_, rest, ok := splitStartLine(input.Data)
	if !ok {
		return &gen.ParseRequestOutput{Error: "no line terminator found; not a valid HTTP message"}, nil
	}
	hs, herr := scanOrderedHeaders(rest)
	if herr != nil {
		return &gen.ParseRequestOutput{Error: herr.Error()}, nil
	}
	if hasAmbiguousFraming(hs) {
		return &gen.ParseRequestOutput{Error: "ambiguous message framing: both Content-Length and Transfer-Encoding present"}, nil
	}

	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(input.Data)))
	if err != nil {
		return &gen.ParseRequestOutput{Error: "parse error: " + err.Error()}, nil
	}

	body, truncated, berr := readBodyBounded(req.Body, clampMaxBody(input.MaxBodyBytes))
	if berr != nil {
		return &gen.ParseRequestOutput{Error: "body read error: " + berr.Error()}, nil
	}

	return &gen.ParseRequestOutput{
		Ok:            true,
		Method:        req.Method,
		Target:        req.RequestURI,
		Version:       req.Proto,
		Headers:       headersToProto(hs),
		Body:          body,
		BodyTruncated: truncated,
	}, nil
}
