package nodes_test

import (
	"context"
	"testing"

	gen "christiangeorgelucas/http-message-tools/gen"
	"christiangeorgelucas/http-message-tools/nodes"
)

// TestParseHeaderBlock_OrderAndRepeats is a hand-computed oracle over a raw
// header block with no start line — order, case-as-written, and repeats
// must all survive exactly as written.
func TestParseHeaderBlock_OrderAndRepeats(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	raw := "Host: example.com\r\nAccept: text/html\r\nAccept: application/json\r\n\r\n"

	got, err := nodes.ParseHeaderBlock(ctx, ax, &gen.ParseHeaderBlockInput{Data: []byte(raw)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok {
		t.Fatalf("expected ok=true, got error %q", got.Error)
	}
	want := []struct{ name, value string }{
		{"Host", "example.com"},
		{"Accept", "text/html"},
		{"Accept", "application/json"},
	}
	if len(got.Headers) != len(want) {
		t.Fatalf("got %d headers, want %d: %+v", len(got.Headers), len(want), got.Headers)
	}
	for i, w := range want {
		if got.Headers[i].Name != w.name || got.Headers[i].Value != w.value {
			t.Errorf("header[%d] = %q:%q, want %q:%q", i, got.Headers[i].Name, got.Headers[i].Value, w.name, w.value)
		}
	}
}

// TestParseHeaderBlock_ObsoleteFolding proves RFC 7230 obsolete
// line-folding (a continuation line beginning with a space/tab) is
// honored — the folded value joins into one logical header value.
func TestParseHeaderBlock_ObsoleteFolding(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	raw := "X-Long: first part\r\n second part\r\n\r\n"

	got, err := nodes.ParseHeaderBlock(ctx, ax, &gen.ParseHeaderBlockInput{Data: []byte(raw)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok || len(got.Headers) != 1 {
		t.Fatalf("got %+v", got)
	}
	if got.Headers[0].Value != "first part second part" {
		t.Errorf("Value = %q, want \"first part second part\"", got.Headers[0].Value)
	}
}

func TestParseHeaderBlock_MissingColonRejected(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ParseHeaderBlock(ctx, ax, &gen.ParseHeaderBlockInput{Data: []byte("NotAHeaderLine\r\n\r\n")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for a line with no colon, got %+v", got)
	}
}

func TestParseHeaderBlock_Empty(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ParseHeaderBlock(ctx, ax, &gen.ParseHeaderBlockInput{Data: []byte("\r\n")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok || len(got.Headers) != 0 {
		t.Fatalf("got %+v", got)
	}
}
