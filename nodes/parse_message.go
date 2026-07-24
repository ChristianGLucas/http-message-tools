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

// Parse a raw HTTP/1.x message of UNKNOWN kind: inspects the first line
// (a response's status-line always starts with "HTTP/") to decide whether
// to parse it as a request or a response, then applies the same logic as
// ParseRequest / ParseResponse. Set request_method="HEAD" if the message
// might be a HEAD response (see ParseResponseInput.request_method).
func ParseMessage(ctx context.Context, ax axiom.Context, input *gen.ParseMessageInput) (*gen.ParseMessageOutput, error) {
	if input == nil || len(input.Data) == 0 {
		return &gen.ParseMessageOutput{Error: "data is required"}, nil
	}
	startLine, rest, ok := splitStartLine(input.Data)
	if !ok {
		return &gen.ParseMessageOutput{Error: "no line terminator found; not a valid HTTP message"}, nil
	}
	hs, herr := scanOrderedHeaders(rest)
	if herr != nil {
		return &gen.ParseMessageOutput{Error: herr.Error()}, nil
	}
	if hasAmbiguousFraming(hs) {
		return &gen.ParseMessageOutput{Error: "ambiguous message framing: both Content-Length and Transfer-Encoding present"}, nil
	}

	isResponse := strings.HasPrefix(startLine, "HTTP/")

	if isResponse {
		var req *http.Request
		if input.RequestMethod != "" {
			req = &http.Request{Method: strings.ToUpper(input.RequestMethod)}
		}
		resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(input.Data)), req)
		if err != nil {
			return &gen.ParseMessageOutput{Error: "parse error: " + err.Error()}, nil
		}
		body, truncated, berr := readBodyBounded(resp.Body, clampMaxBody(input.MaxBodyBytes))
		if berr != nil {
			return &gen.ParseMessageOutput{Error: "body read error: " + berr.Error()}, nil
		}
		return &gen.ParseMessageOutput{
			Ok:            true,
			IsRequest:     false,
			Version:       resp.Proto,
			StatusCode:    int32(resp.StatusCode),
			Reason:        strings.TrimSpace(strings.TrimPrefix(resp.Status, strconv.Itoa(resp.StatusCode))),
			Headers:       headersToProto(hs),
			Body:          body,
			BodyTruncated: truncated,
		}, nil
	}

	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(input.Data)))
	if err != nil {
		return &gen.ParseMessageOutput{Error: "parse error: " + err.Error()}, nil
	}
	body, truncated, berr := readBodyBounded(req.Body, clampMaxBody(input.MaxBodyBytes))
	if berr != nil {
		return &gen.ParseMessageOutput{Error: "body read error: " + berr.Error()}, nil
	}
	return &gen.ParseMessageOutput{
		Ok:            true,
		IsRequest:     true,
		Method:        req.Method,
		Target:        req.RequestURI,
		Version:       req.Proto,
		Headers:       headersToProto(hs),
		Body:          body,
		BodyTruncated: truncated,
	}, nil
}
