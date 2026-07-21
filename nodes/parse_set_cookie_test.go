package nodes_test

import (
	"context"
	"testing"

	gen "christiangeorgelucas/http-message-tools/gen"
	"christiangeorgelucas/http-message-tools/nodes"
)

// TestParseSetCookie_FullAttributes is a hand-computed oracle: every
// attribute in the input is independently checkable against the literal
// string.
func TestParseSetCookie_FullAttributes(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	raw := "session=abc123; Domain=example.com; Path=/; Expires=Wed, 21 Oct 2026 07:28:00 GMT; Max-Age=3600; Secure; HttpOnly; SameSite=Strict"

	got, err := nodes.ParseSetCookie(ctx, ax, &gen.ParseSetCookieInput{Value: raw})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok {
		t.Fatalf("expected ok=true, got error %q", got.Error)
	}
	if got.Name != "session" || got.Value != "abc123" {
		t.Errorf("Name/Value = %q/%q, want session/abc123", got.Name, got.Value)
	}
	if got.Domain != "example.com" {
		t.Errorf("Domain = %q, want example.com", got.Domain)
	}
	if got.Path != "/" {
		t.Errorf("Path = %q, want /", got.Path)
	}
	if got.Expires != "Wed, 21 Oct 2026 07:28:00 GMT" {
		t.Errorf("Expires = %q, want Wed, 21 Oct 2026 07:28:00 GMT", got.Expires)
	}
	if !got.HasMaxAge || got.MaxAge != 3600 {
		t.Errorf("MaxAge/HasMaxAge = %d/%v, want 3600/true", got.MaxAge, got.HasMaxAge)
	}
	if !got.Secure {
		t.Errorf("Secure = false, want true")
	}
	if !got.HttpOnly {
		t.Errorf("HttpOnly = false, want true")
	}
	if got.SameSite != "Strict" {
		t.Errorf("SameSite = %q, want Strict", got.SameSite)
	}
}

func TestParseSetCookie_Minimal(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ParseSetCookie(ctx, ax, &gen.ParseSetCookieInput{Value: "a=b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok || got.Name != "a" || got.Value != "b" {
		t.Fatalf("got %+v", got)
	}
	if got.HasMaxAge {
		t.Errorf("HasMaxAge = true, want false for a cookie with no Max-Age attribute")
	}
	if got.Secure || got.HttpOnly {
		t.Errorf("Secure/HttpOnly should default false, got %v/%v", got.Secure, got.HttpOnly)
	}
	if got.SameSite != "" {
		t.Errorf("SameSite = %q, want empty", got.SameSite)
	}
}

func TestParseSetCookie_EmptyValueRejected(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ParseSetCookie(ctx, ax, &gen.ParseSetCookieInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for empty value")
	}
}
