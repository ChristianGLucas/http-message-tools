package nodes_test

import (
	"context"
	"testing"

	gen "christiangeorgelucas/http-message-tools/gen"
	"christiangeorgelucas/http-message-tools/nodes"
)

func TestParseMessage_DetectsRequest(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	raw := "GET /a HTTP/1.1\r\nHost: x\r\n\r\n"

	got, err := nodes.ParseMessage(ctx, ax, &gen.ParseMessageInput{Data: []byte(raw)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok || !got.IsRequest || got.Method != "GET" || got.Target != "/a" {
		t.Fatalf("got %+v", got)
	}
}

func TestParseMessage_DetectsResponse(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	raw := "HTTP/1.1 500 Internal Server Error\r\nContent-Length: 0\r\n\r\n"

	got, err := nodes.ParseMessage(ctx, ax, &gen.ParseMessageInput{Data: []byte(raw)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok || got.IsRequest || got.StatusCode != 500 || got.Reason != "Internal Server Error" {
		t.Fatalf("got %+v", got)
	}
}

func TestParseMessage_Malformed(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ParseMessage(ctx, ax, &gen.ParseMessageInput{Data: []byte("garbage\r\n\r\n")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for garbage input, got %+v", got)
	}
}
