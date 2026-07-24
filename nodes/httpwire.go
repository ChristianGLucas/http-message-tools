package nodes

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"

	gen "christiangeorgelucas/http-message-tools/gen"
)

// Payload size, header count, and line-length limits are enforced by the
// Axiom platform (ingress, transport, and sandboxed execution) — this
// package does not impose its own. defaultMaxBody is not a rejection bound;
// it's just the default (and ceiling) clampMaxBody/clampMaxOutput apply when
// a caller doesn't specify how many body bytes they want read.
const (
	defaultMaxBody = 4 << 20 // 4 MiB
)

var httpVersionRe = regexp.MustCompile(`^HTTP/[0-9]\.[0-9]$`)

// header is the package-internal, order-preserving representation used
// while scanning; headersToProto converts a slice of these into the wire
// protobuf repeated field.
type header struct {
	name  string
	value string
}

func headersToProto(hs []header) []*gen.HttpHeader {
	out := make([]*gen.HttpHeader, 0, len(hs))
	for _, h := range hs {
		out = append(out, &gen.HttpHeader{Name: h.name, Value: h.value})
	}
	return out
}

// splitStartLine finds the message's first line (request-line or
// status-line) and returns it (with any trailing \r stripped) plus
// everything after the line terminator.
func splitStartLine(data []byte) (line string, rest []byte, ok bool) {
	i := bytes.IndexByte(data, '\n')
	if i < 0 {
		return "", nil, false
	}
	l := data[:i]
	l = bytes.TrimSuffix(l, []byte("\r"))
	return string(l), data[i+1:], true
}

// scanOrderedHeaders reads "Name: value" lines (honoring RFC 7230 obsolete
// line-folding via textproto.Reader.ReadContinuedLineBytes, which owns the
// actual fold-continuation grammar) from data until a blank line or EOF,
// preserving both wire ORDER and REPEATS — something net/http and
// net/textproto's own map-based header types cannot do, since Go map
// iteration order is not the wire order.
func scanOrderedHeaders(data []byte) ([]header, error) {
	tp := textproto.NewReader(bufio.NewReader(bytes.NewReader(data)))
	var hs []header
	for {
		lineBytes, err := tp.ReadContinuedLineBytes()
		if err != nil {
			// EOF with no blank line seen just means the input ended right
			// at (or before) the header/body boundary; return what we have.
			break
		}
		if len(lineBytes) == 0 {
			break // blank line: end of header block
		}
		idx := bytes.IndexByte(lineBytes, ':')
		if idx < 0 {
			return hs, fmt.Errorf("malformed header line %q: missing ':'", string(lineBytes))
		}
		name := string(bytes.TrimSpace(lineBytes[:idx]))
		value := string(bytes.TrimSpace(lineBytes[idx+1:]))
		if name == "" {
			return hs, fmt.Errorf("malformed header line %q: empty name", string(lineBytes))
		}
		hs = append(hs, header{name: name, value: value})
	}
	return hs, nil
}

// scanHeadersFlexible is scanOrderedHeaders with a fallback: if data does
// not parse as a bare header block (e.g. because its first line is a
// request-line or status-line with no colon), it strips the first line and
// retries on the remainder. Used by GetHeader so it accepts either a bare
// header block or a complete message.
func scanHeadersFlexible(data []byte) ([]header, error) {
	if hs, err := scanOrderedHeaders(data); err == nil {
		return hs, nil
	}
	if _, rest, ok := splitStartLine(data); ok {
		if hs, err := scanOrderedHeaders(rest); err == nil {
			return hs, nil
		}
	}
	// Fall through to the original (block-mode) error, which is usually
	// the more informative one.
	return scanOrderedHeaders(data)
}

// hasAmbiguousFraming reports whether both Content-Length and
// Transfer-Encoding are present (case-insensitive header names) — the
// classic HTTP request-smuggling ambiguity. net/http silently resolves
// this in favor of chunked and DROPS both headers from the returned map
// rather than erroring (verified against go1.25), so callers relying on
// net/http alone never see the ambiguity. We reject it explicitly instead
// of guessing which framing a downstream proxy would honor.
func hasAmbiguousFraming(hs []header) bool {
	hasCL, hasTE := false, false
	for _, h := range hs {
		switch strings.ToLower(h.name) {
		case "content-length":
			hasCL = true
		case "transfer-encoding":
			hasTE = true
		}
	}
	return hasCL && hasTE
}

// readBodyBounded reads at most max bytes from r. If more than max bytes
// were available it returns the first max bytes with truncated=true. A
// genuine read error (e.g. the message declared a Content-Length longer
// than the bytes actually supplied, which surfaces as io.ErrUnexpectedEOF)
// is propagated rather than silently swallowed, since that indicates the
// input was not actually a complete, valid message.
func readBodyBounded(r io.Reader, max int) (data []byte, truncated bool, err error) {
	if r == nil {
		return nil, false, nil
	}
	lr := io.LimitReader(r, int64(max)+1)
	b, err := io.ReadAll(lr)
	if err != nil {
		return nil, false, err
	}
	if len(b) > max {
		return b[:max], true, nil
	}
	return b, false, nil
}

func clampMaxBody(requested int32) int {
	if requested <= 0 || int(requested) > defaultMaxBody {
		return defaultMaxBody
	}
	return int(requested)
}

func clampMaxOutput(requested int32) int {
	if requested <= 0 || int(requested) > defaultMaxBody {
		return defaultMaxBody
	}
	return int(requested)
}

// --- token / field-value validation for the Build* nodes -------------------
//
// A wire builder that concatenates caller-supplied strings into header
// lines is a classic CRLF-header-injection surface: if a "value" is allowed
// to contain \r\n, a caller can smuggle extra headers or even a second
// message into the output. RFC 7230 already defines the grammar a
// conforming name/value must satisfy (token / field-content); we enforce
// exactly that grammar here, rejecting anything that would corrupt the
// message we are about to emit, rather than escaping or best-effort
// sanitizing it.

func isTokenChar(b byte) bool {
	switch {
	case b >= 'a' && b <= 'z', b >= 'A' && b <= 'Z', b >= '0' && b <= '9':
		return true
	}
	switch b {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	return false
}

func isValidToken(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !isTokenChar(s[i]) {
			return false
		}
	}
	return true
}

// isValidFieldValue rejects control characters (which would let a caller
// inject a line break, and hence a whole new header or start-line, into
// the message we build) while allowing ordinary printable text, spaces,
// tabs, and obs-text (0x80-0xFF) per RFC 7230's field-content grammar.
func isValidFieldValue(s string) bool {
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b == '\t' {
			continue
		}
		if b < 0x20 || b == 0x7f {
			return false
		}
	}
	return true
}

// isValidTargetOrVersion rejects control characters and whitespace that
// would split a request-line/status-line into more than the intended
// fields (or inject a second line) when we assemble raw wire bytes.
func isValidTargetOrVersion(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b <= 0x20 || b == 0x7f {
			return false
		}
	}
	return true
}

func normalizeVersion(v string) string {
	if v == "" {
		return "HTTP/1.1"
	}
	return v
}

func validVersion(v string) bool {
	return httpVersionRe.MatchString(v)
}

// buildHeaderErrors validates every header name/value the caller supplied
// for use in a Build* node and returns the first problem found, or "".
func validateHeadersForBuild(hs []*gen.HttpHeader) error {
	for _, h := range hs {
		if !isValidToken(h.Name) {
			return fmt.Errorf("invalid header name %q: must be an RFC 7230 token", h.Name)
		}
		if !isValidFieldValue(h.Value) {
			return fmt.Errorf("invalid header value for %q: contains a control character", h.Name)
		}
	}
	return nil
}

func writeHeaders(buf *bytes.Buffer, hs []*gen.HttpHeader) {
	for _, h := range hs {
		buf.WriteString(h.Name)
		buf.WriteString(": ")
		buf.WriteString(h.Value)
		buf.WriteString("\r\n")
	}
}

// statusText mirrors http.StatusText but is exposed here for readability
// at call sites; kept as a thin wrapper (not a reimplementation) around
// the standard library's own table.
func statusText(code int) string {
	return http.StatusText(code)
}

func parseInt32(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 32)
}
