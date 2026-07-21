package nodes_test

import (
	"context"
	"testing"

	gen "christiangeorgelucas/http-message-tools/gen"
	"christiangeorgelucas/http-message-tools/nodes"
)

func TestParseCookieHeader_Golden(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ParseCookieHeader(ctx, ax, &gen.ParseCookieHeaderInput{Value: "a=1; b=2; c=3"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok {
		t.Fatalf("expected ok=true, got error %q", got.Error)
	}
	want := map[string]string{"a": "1", "b": "2", "c": "3"}
	if len(got.Cookies) != 3 {
		t.Fatalf("got %d cookies, want 3: %+v", len(got.Cookies), got.Cookies)
	}
	for _, c := range got.Cookies {
		if want[c.Name] != c.Value {
			t.Errorf("cookie %q = %q, want %q", c.Name, c.Value, want[c.Name])
		}
	}
}

func TestParseCookieHeader_SingleCookie(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ParseCookieHeader(ctx, ax, &gen.ParseCookieHeaderInput{Value: "session=abc123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok || len(got.Cookies) != 1 || got.Cookies[0].Name != "session" || got.Cookies[0].Value != "abc123" {
		t.Fatalf("got %+v", got)
	}
}

func TestParseCookieHeader_EmptyValueRejected(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ParseCookieHeader(ctx, ax, &gen.ParseCookieHeaderInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for empty value")
	}
}
