package nodes_test

import (
	"context"
	"testing"

	gen "christiangeorgelucas/http-message-tools/gen"
	"christiangeorgelucas/http-message-tools/nodes"
)

// TestBuildRequest_ExactBytes is a hand-computed oracle: the exact wire
// bytes are known directly from the structured input.
func TestBuildRequest_ExactBytes(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.BuildRequest(ctx, ax, &gen.BuildRequestInput{
		Method: "POST",
		Target: "/submit",
		Headers: []*gen.HttpHeader{
			{Name: "Host", Value: "example.com"},
			{Name: "Content-Type", Value: "text/plain"},
		},
		Body: []byte("hi"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok {
		t.Fatalf("expected ok=true, got error %q", got.Error)
	}
	want := "POST /submit HTTP/1.1\r\nHost: example.com\r\nContent-Type: text/plain\r\n\r\nhi"
	if string(got.Data) != want {
		t.Errorf("Data = %q, want %q", got.Data, want)
	}
}

func TestBuildRequest_DefaultsTargetAndVersion(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.BuildRequest(ctx, ax, &gen.BuildRequestInput{Method: "GET"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok {
		t.Fatalf("expected ok=true, got error %q", got.Error)
	}
	want := "GET / HTTP/1.1\r\n\r\n"
	if string(got.Data) != want {
		t.Errorf("Data = %q, want %q", got.Data, want)
	}
}

// TestBuildRequest_RoundTrip proves Build then Parse recovers the original
// structured fields — the strongest available correctness check for a
// builder, since it is checked against the package's own parser rather
// than merely asserting the bytes look plausible.
func TestBuildRequest_RoundTrip(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	built, err := nodes.BuildRequest(ctx, ax, &gen.BuildRequestInput{
		Method:  "PUT",
		Target:  "/x/y?z=1",
		Version: "HTTP/1.1",
		Headers: []*gen.HttpHeader{{Name: "Host", Value: "a.example"}, {Name: "Content-Length", Value: "3"}},
		Body:    []byte("abc"),
	})
	if err != nil || !built.Ok {
		t.Fatalf("build failed: err=%v out=%+v", err, built)
	}

	parsed, err := nodes.ParseRequest(ctx, ax, &gen.ParseRequestInput{Data: built.Data})
	if err != nil || !parsed.Ok {
		t.Fatalf("round-trip parse failed: err=%v out=%+v", err, parsed)
	}
	if parsed.Method != "PUT" || parsed.Target != "/x/y?z=1" || string(parsed.Body) != "abc" {
		t.Errorf("round-trip mismatch: %+v", parsed)
	}
}

// TestBuildRequest_HeaderInjectionRejected proves a caller cannot smuggle
// an extra header (or a whole second message) via a CRLF embedded in a
// header value.
func TestBuildRequest_HeaderInjectionRejected(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.BuildRequest(ctx, ax, &gen.BuildRequestInput{
		Method:  "GET",
		Target:  "/",
		Headers: []*gen.HttpHeader{{Name: "X-Evil", Value: "a\r\nX-Injected: yes"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for a header value containing CRLF, got %+v", got)
	}
}

// TestBuildRequest_MismatchedContentLengthRejected is the killing
// regression test for a CRITICAL an independent adversarial review found:
// http.ReadRequest only parses the start-line/headers eagerly and leaves
// the body as a lazily-read io.ReadCloser, so an earlier version of the
// self-validate step (which called ReadRequest but never read the body)
// let a Content-Length that does not match the actual body length through
// as ok=true — even though the package's OWN ParseRequest would reject the
// exact same bytes with a body-read error. A realistic caller-supplied
// mismatch (e.g. a hand-typed or stale Content-Length) must be rejected.
func TestBuildRequest_MismatchedContentLengthRejected(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.BuildRequest(ctx, ax, &gen.BuildRequestInput{
		Method:  "POST",
		Target:  "/",
		Headers: []*gen.HttpHeader{{Name: "Content-Length", Value: "100"}},
		Body:    []byte("hi"), // only 2 bytes, not 100
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for a Content-Length that does not match the actual body length, got %+v", got)
	}
	if got.Error == "" {
		t.Errorf("expected a structured error message, got empty string")
	}
}

func TestBuildRequest_InvalidHeaderNameRejected(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.BuildRequest(ctx, ax, &gen.BuildRequestInput{
		Method:  "GET",
		Target:  "/",
		Headers: []*gen.HttpHeader{{Name: "Bad Name:", Value: "x"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for an invalid header name, got %+v", got)
	}
}

func TestBuildRequest_MethodRequired(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.BuildRequest(ctx, ax, &gen.BuildRequestInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false when method is empty")
	}
}
