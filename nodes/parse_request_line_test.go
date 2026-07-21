package nodes_test

import (
	"context"
	"testing"

	gen "christiangeorgelucas/http-message-tools/gen"
	"christiangeorgelucas/http-message-tools/nodes"
)

func TestParseRequestLine_Golden(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ParseRequestLine(ctx, ax, &gen.ParseRequestLineInput{Line: "GET /path?q=1 HTTP/1.1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok || got.Method != "GET" || got.Target != "/path?q=1" || got.Version != "HTTP/1.1" {
		t.Fatalf("got %+v", got)
	}
}

func TestParseRequestLine_EmbeddedNewlineRejected(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ParseRequestLine(ctx, ax, &gen.ParseRequestLineInput{Line: "GET / HTTP/1.1\r\nX-Injected: 1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for a line with an embedded newline, got %+v", got)
	}
}

func TestParseRequestLine_Malformed(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ParseRequestLine(ctx, ax, &gen.ParseRequestLineInput{Line: "totally not a request line"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for malformed line, got %+v", got)
	}
}
