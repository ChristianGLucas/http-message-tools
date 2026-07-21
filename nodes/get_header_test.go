package nodes_test

import (
	"context"
	"testing"

	gen "christiangeorgelucas/http-message-tools/gen"
	"christiangeorgelucas/http-message-tools/nodes"
)

func TestGetHeader_CaseInsensitive(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	raw := "Content-Type: text/html\r\nHost: a\r\n\r\n"

	got, err := nodes.GetHeader(ctx, ax, &gen.GetHeaderInput{Data: []byte(raw), Name: "content-type"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok || !got.Found || got.FirstValue != "text/html" {
		t.Fatalf("got %+v", got)
	}
}

// TestGetHeader_RepeatedHeader proves multiple occurrences (e.g.
// Set-Cookie) are all returned, in order.
func TestGetHeader_RepeatedHeader(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	raw := "Set-Cookie: a=1\r\nSet-Cookie: b=2\r\n\r\n"

	got, err := nodes.GetHeader(ctx, ax, &gen.GetHeaderInput{Data: []byte(raw), Name: "Set-Cookie"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Found || len(got.Values) != 2 || got.Values[0] != "a=1" || got.Values[1] != "b=2" {
		t.Fatalf("got %+v", got)
	}
	if got.FirstValue != "a=1" {
		t.Errorf("FirstValue = %q, want a=1", got.FirstValue)
	}
}

func TestGetHeader_NotFound(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	raw := "Host: a\r\n\r\n"

	got, err := nodes.GetHeader(ctx, ax, &gen.GetHeaderInput{Data: []byte(raw), Name: "X-Missing"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok || got.Found {
		t.Fatalf("expected ok=true, found=false, got %+v", got)
	}
}

// TestGetHeader_FullMessageFallback proves GetHeader accepts a COMPLETE
// message (with a leading request-line), not just a bare header block —
// the request-line is detected and skipped.
func TestGetHeader_FullMessageFallback(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	raw := "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"

	got, err := nodes.GetHeader(ctx, ax, &gen.GetHeaderInput{Data: []byte(raw), Name: "Host"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Found || got.FirstValue != "example.com" {
		t.Fatalf("got %+v", got)
	}
}

func TestGetHeader_EmptyNameRejected(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.GetHeader(ctx, ax, &gen.GetHeaderInput{Data: []byte("Host: a\r\n\r\n")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false when name is empty, got %+v", got)
	}
}
