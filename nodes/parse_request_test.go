package nodes_test

import (
	"context"
	"strconv"
	"testing"

	gen "christiangeorgelucas/http-message-tools/gen"
	"christiangeorgelucas/http-message-tools/nodes"
)

// TestParseRequest_Golden is a hand-computed oracle: the expected
// method/target/version/headers/body are known directly from the literal
// wire text below, independent of the implementation under test.
func TestParseRequest_Golden(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	raw := "POST /submit?x=1 HTTP/1.1\r\n" +
		"Host: example.com\r\n" +
		"X-Custom: first\r\n" +
		"Content-Type: text/plain\r\n" +
		"X-Custom: second\r\n" +
		"Content-Length: 5\r\n" +
		"\r\n" +
		"hello"

	got, err := nodes.ParseRequest(ctx, ax, &gen.ParseRequestInput{Data: []byte(raw)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok {
		t.Fatalf("expected ok=true, got error %q", got.Error)
	}
	if got.Method != "POST" {
		t.Errorf("Method = %q, want POST", got.Method)
	}
	if got.Target != "/submit?x=1" {
		t.Errorf("Target = %q, want /submit?x=1", got.Target)
	}
	if got.Version != "HTTP/1.1" {
		t.Errorf("Version = %q, want HTTP/1.1", got.Version)
	}
	if string(got.Body) != "hello" {
		t.Errorf("Body = %q, want hello", got.Body)
	}
	if got.BodyTruncated {
		t.Errorf("BodyTruncated = true, want false")
	}
	// Wire order AND the repeated X-Custom header must both survive.
	wantNames := []string{"Host", "X-Custom", "Content-Type", "X-Custom", "Content-Length"}
	if len(got.Headers) != len(wantNames) {
		t.Fatalf("got %d headers, want %d: %+v", len(got.Headers), len(wantNames), got.Headers)
	}
	for i, name := range wantNames {
		if got.Headers[i].Name != name {
			t.Errorf("header[%d].Name = %q, want %q", i, got.Headers[i].Name, name)
		}
	}
	if got.Headers[1].Value != "first" || got.Headers[3].Value != "second" {
		t.Errorf("X-Custom values = %q, %q, want first, second", got.Headers[1].Value, got.Headers[3].Value)
	}
}

func TestParseRequest_NoBodyGET(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	raw := "GET /health HTTP/1.1\r\nHost: a\r\n\r\n"

	got, err := nodes.ParseRequest(ctx, ax, &gen.ParseRequestInput{Data: []byte(raw)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok || got.Method != "GET" || len(got.Body) != 0 {
		t.Fatalf("got %+v", got)
	}
}

// TestParseRequest_AmbiguousFraming proves the request-smuggling guard: a
// realistic caller-reachable message with BOTH Content-Length and
// Transfer-Encoding is rejected rather than silently resolved one way (as
// net/http itself does — verified separately against go1.25: it drops both
// headers from its Header map and picks chunked without erroring).
func TestParseRequest_AmbiguousFraming(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	raw := "POST /x HTTP/1.1\r\nHost: a\r\nContent-Length: 5\r\nTransfer-Encoding: chunked\r\n\r\n5\r\nhello\r\n0\r\n\r\n"

	got, err := nodes.ParseRequest(ctx, ax, &gen.ParseRequestInput{Data: []byte(raw)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for ambiguous framing, got %+v", got)
	}
	if got.Error == "" {
		t.Errorf("expected a structured error message, got empty string")
	}
}

func TestParseRequest_Malformed(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ParseRequest(ctx, ax, &gen.ParseRequestInput{Data: []byte("NOT AN HTTP REQUEST\r\n\r\n")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for malformed input, got %+v", got)
	}
}

func TestParseRequest_EmptyDataRejected(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ParseRequest(ctx, ax, &gen.ParseRequestInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok || got.Error == "" {
		t.Fatalf("expected structured error for empty input, got %+v", got)
	}
}

// TestParseRequest_LargeInputParses proves this node imposes no self-sized
// total-input cap: a well-formed request whose raw bytes exceed the old
// 4 MiB package-level reject threshold must still parse (not be rejected
// outright with a "size limit" error). Payload-size limiting is the Axiom
// platform's job (ingress, transport, sandboxed execution), not this
// node's — so no input length, however large, should trigger got.Ok=false
// here. (The body itself may still come back BodyTruncated=true: that's the
// unrelated, still-present default max-body-read behavior, not a rejection.)
func TestParseRequest_LargeInputParses(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	body := make([]byte, (4<<20)+1024) // > the old 4 MiB total-input bound
	for i := range body {
		body[i] = 'a'
	}
	raw := []byte("POST /upload HTTP/1.1\r\nHost: a\r\nContent-Length: " +
		strconv.Itoa(len(body)) + "\r\n\r\n")
	raw = append(raw, body...)

	got, err := nodes.ParseRequest(ctx, ax, &gen.ParseRequestInput{Data: raw})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok {
		t.Fatalf("expected ok=true for a large but well-formed request, got error %q", got.Error)
	}
	if got.Method != "POST" || got.Target != "/upload" {
		t.Errorf("got Method=%q Target=%q, want POST /upload", got.Method, got.Target)
	}
}
