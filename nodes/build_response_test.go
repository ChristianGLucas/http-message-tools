package nodes_test

import (
	"context"
	"testing"

	gen "christiangeorgelucas/http-message-tools/gen"
	"christiangeorgelucas/http-message-tools/nodes"
)

func TestBuildResponse_ExactBytes(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.BuildResponse(ctx, ax, &gen.BuildResponseInput{
		StatusCode: 200,
		Headers:    []*gen.HttpHeader{{Name: "Content-Type", Value: "text/plain"}},
		Body:       []byte("hi"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok {
		t.Fatalf("expected ok=true, got error %q", got.Error)
	}
	want := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nhi"
	if string(got.Data) != want {
		t.Errorf("Data = %q, want %q", got.Data, want)
	}
}

// TestBuildResponse_DefaultReasonFromStatusText is an independent oracle:
// the expected reason phrase for 404 is looked up from the RFC 9110 table
// (net/http.StatusText), a value known independently of BuildResponse's
// own logic.
func TestBuildResponse_DefaultReasonFromStatusText(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.BuildResponse(ctx, ax, &gen.BuildResponseInput{StatusCode: 404})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok {
		t.Fatalf("expected ok=true, got error %q", got.Error)
	}
	want := "HTTP/1.1 404 Not Found\r\n\r\n"
	if string(got.Data) != want {
		t.Errorf("Data = %q, want %q", got.Data, want)
	}
}

func TestBuildResponse_ExplicitReasonOverridesDefault(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.BuildResponse(ctx, ax, &gen.BuildResponseInput{StatusCode: 200, Reason: "Custom"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok {
		t.Fatalf("expected ok=true, got error %q", got.Error)
	}
	want := "HTTP/1.1 200 Custom\r\n\r\n"
	if string(got.Data) != want {
		t.Errorf("Data = %q, want %q", got.Data, want)
	}
}

// TestBuildResponse_UnknownCodeRequiresExplicitReason proves the
// no-standard-reason-phrase path is actually enforced (599 has no entry
// in net/http.StatusText's table).
func TestBuildResponse_UnknownCodeRequiresExplicitReason(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.BuildResponse(ctx, ax, &gen.BuildResponseInput{StatusCode: 599})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for an unrecognized status code with no explicit reason")
	}

	got2, err := nodes.BuildResponse(ctx, ax, &gen.BuildResponseInput{StatusCode: 599, Reason: "Custom Code"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got2.Ok {
		t.Fatalf("expected ok=true when reason is supplied explicitly, got error %q", got2.Error)
	}
}

func TestBuildResponse_RoundTrip(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	built, err := nodes.BuildResponse(ctx, ax, &gen.BuildResponseInput{
		StatusCode: 201,
		Headers:    []*gen.HttpHeader{{Name: "Content-Length", Value: "2"}},
		Body:       []byte("ok"),
	})
	if err != nil || !built.Ok {
		t.Fatalf("build failed: err=%v out=%+v", err, built)
	}

	parsed, err := nodes.ParseResponse(ctx, ax, &gen.ParseResponseInput{Data: built.Data})
	if err != nil || !parsed.Ok {
		t.Fatalf("round-trip parse failed: err=%v out=%+v", err, parsed)
	}
	if parsed.StatusCode != 201 || parsed.Reason != "Created" || string(parsed.Body) != "ok" {
		t.Errorf("round-trip mismatch: %+v", parsed)
	}
}

func TestBuildResponse_HeaderInjectionRejected(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.BuildResponse(ctx, ax, &gen.BuildResponseInput{
		StatusCode: 200,
		Headers:    []*gen.HttpHeader{{Name: "X-Evil", Value: "a\r\nX-Injected: yes"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for a header value containing CRLF, got %+v", got)
	}
}

func TestBuildResponse_InvalidStatusCodeRejected(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.BuildResponse(ctx, ax, &gen.BuildResponseInput{StatusCode: 999})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for an out-of-range status code")
	}
}
