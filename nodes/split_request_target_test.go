package nodes_test

import (
	"context"
	"testing"

	gen "christiangeorgelucas/http-message-tools/gen"
	"christiangeorgelucas/http-message-tools/nodes"
)

func TestSplitRequestTarget_WithQuery(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.SplitRequestTarget(ctx, ax, &gen.SplitRequestTargetInput{Target: "/search?q=go+http&page=2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok || got.Path != "/search" || got.Query != "q=go+http&page=2" {
		t.Fatalf("got %+v", got)
	}
}

func TestSplitRequestTarget_NoQuery(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.SplitRequestTarget(ctx, ax, &gen.SplitRequestTargetInput{Target: "/plain/path"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok || got.Path != "/plain/path" || got.Query != "" {
		t.Fatalf("got %+v", got)
	}
}

func TestSplitRequestTarget_AsteriskForm(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.SplitRequestTarget(ctx, ax, &gen.SplitRequestTargetInput{Target: "*"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok || got.Path != "*" || got.Query != "" {
		t.Fatalf("got %+v", got)
	}
}

func TestSplitRequestTarget_EmptyTrailingQuery(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.SplitRequestTarget(ctx, ax, &gen.SplitRequestTargetInput{Target: "/a?"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok || got.Path != "/a" || got.Query != "" {
		t.Fatalf("got %+v", got)
	}
}

func TestSplitRequestTarget_EmptyTargetRejected(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.SplitRequestTarget(ctx, ax, &gen.SplitRequestTargetInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for empty target")
	}
}
