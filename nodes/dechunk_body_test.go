package nodes_test

import (
	"context"
	"testing"

	gen "christiangeorgelucas/http-message-tools/gen"
	"christiangeorgelucas/http-message-tools/nodes"
)

// TestDechunkBody_Golden is the canonical RFC 7230 §4.1 example: decoding
// "4\r\nWiki\r\n5\r\npedia\r\n0\r\n\r\n" must yield exactly "Wikipedia".
func TestDechunkBody_Golden(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	raw := "4\r\nWiki\r\n5\r\npedia\r\n0\r\n\r\n"

	got, err := nodes.DechunkBody(ctx, ax, &gen.DechunkBodyInput{Data: []byte(raw)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok {
		t.Fatalf("expected ok=true, got error %q", got.Error)
	}
	if string(got.Data) != "Wikipedia" {
		t.Errorf("Data = %q, want Wikipedia", got.Data)
	}
	if got.Truncated {
		t.Errorf("Truncated = true, want false")
	}
}

func TestDechunkBody_MalformedChunkSize(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.DechunkBody(ctx, ax, &gen.DechunkBodyInput{Data: []byte("ZZZ\r\nhello\r\n")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for malformed chunk size, got %+v", got)
	}
}

// TestDechunkBody_OutputBounded proves a declared chunk size larger than
// the caller's requested cap is truncated rather than fully allocated.
func TestDechunkBody_OutputBounded(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	raw := "9\r\n123456789\r\n0\r\n\r\n" // 9-byte chunk

	got, err := nodes.DechunkBody(ctx, ax, &gen.DechunkBodyInput{Data: []byte(raw), MaxOutputBytes: 4})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok {
		t.Fatalf("expected ok=true, got error %q", got.Error)
	}
	if !got.Truncated {
		t.Errorf("expected Truncated=true")
	}
	if string(got.Data) != "1234" {
		t.Errorf("Data = %q, want 1234", got.Data)
	}
}

func TestDechunkBody_EmptyDataRejected(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.DechunkBody(ctx, ax, &gen.DechunkBodyInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for empty input")
	}
}
