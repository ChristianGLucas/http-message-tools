package nodes_test

import (
	"context"
	"testing"

	gen "christiangeorgelucas/http-message-tools/gen"
	"christiangeorgelucas/http-message-tools/nodes"
)

func TestParseStatusLine_Golden(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ParseStatusLine(ctx, ax, &gen.ParseStatusLineInput{Line: "HTTP/1.1 404 Not Found"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok || got.Version != "HTTP/1.1" || got.StatusCode != 404 || got.Reason != "Not Found" {
		t.Fatalf("got %+v", got)
	}
}

func TestParseStatusLine_MultiWordReason(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ParseStatusLine(ctx, ax, &gen.ParseStatusLineInput{Line: "HTTP/1.1 500 Internal Server Error"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok || got.StatusCode != 500 || got.Reason != "Internal Server Error" {
		t.Fatalf("got %+v", got)
	}
}

func TestParseStatusLine_Malformed(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ParseStatusLine(ctx, ax, &gen.ParseStatusLineInput{Line: "not a status line"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for malformed line, got %+v", got)
	}
}
