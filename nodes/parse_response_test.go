package nodes_test

import (
	"context"
	"testing"

	gen "christiangeorgelucas/http-message-tools/gen"
	"christiangeorgelucas/http-message-tools/nodes"
)

func TestParseResponse_Golden(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	raw := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: application/json\r\n" +
		"Set-Cookie: a=1\r\n" +
		"Set-Cookie: b=2\r\n" +
		"Content-Length: 13\r\n" +
		"\r\n" +
		`{"ok":true}` + "\r\n"

	got, err := nodes.ParseResponse(ctx, ax, &gen.ParseResponseInput{Data: []byte(raw)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok {
		t.Fatalf("expected ok=true, got error %q", got.Error)
	}
	if got.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", got.StatusCode)
	}
	if got.Reason != "OK" {
		t.Errorf("Reason = %q, want OK", got.Reason)
	}
	if got.Version != "HTTP/1.1" {
		t.Errorf("Version = %q, want HTTP/1.1", got.Version)
	}
	if string(got.Body) != `{"ok":true}`+"\r\n" {
		t.Errorf("Body = %q", got.Body)
	}
	// Two Set-Cookie occurrences must both survive, in order.
	var cookieVals []string
	for _, h := range got.Headers {
		if h.Name == "Set-Cookie" {
			cookieVals = append(cookieVals, h.Value)
		}
	}
	if len(cookieVals) != 2 || cookieVals[0] != "a=1" || cookieVals[1] != "b=2" {
		t.Errorf("Set-Cookie values = %v, want [a=1 b=2]", cookieVals)
	}
}

func TestParseResponse_404(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	raw := "HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\n\r\n"

	got, err := nodes.ParseResponse(ctx, ax, &gen.ParseResponseInput{Data: []byte(raw)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Ok || got.StatusCode != 404 || got.Reason != "Not Found" {
		t.Fatalf("got %+v", got)
	}
}

// TestParseResponse_HeadResponseContentLength proves the request_method
// field actually changes behavior: a HEAD response legitimately carries a
// Content-Length with NO body bytes present on the wire (RFC 7230 §3.3.3).
// Without request_method="HEAD", net/http.ReadResponse would try to read
// 5 declared body bytes that are not there and fail with an unexpected-EOF
// body-read error; with it, the body is correctly treated as absent.
func TestParseResponse_HeadResponseContentLength(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	raw := "HTTP/1.1 200 OK\r\nContent-Length: 5\r\n\r\n" // no body bytes follow

	gotHead, err := nodes.ParseResponse(ctx, ax, &gen.ParseResponseInput{Data: []byte(raw), RequestMethod: "HEAD"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gotHead.Ok {
		t.Fatalf("expected ok=true for HEAD response, got error %q", gotHead.Error)
	}
	if len(gotHead.Body) != 0 {
		t.Errorf("expected empty body for HEAD response, got %q", gotHead.Body)
	}

	gotGET, err := nodes.ParseResponse(ctx, ax, &gen.ParseResponseInput{Data: []byte(raw)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotGET.Ok {
		t.Fatalf("expected ok=false when the same bytes are read as a GET response (declared body missing), got %+v", gotGET)
	}
}

func TestParseResponse_AmbiguousFraming(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)
	raw := "HTTP/1.1 200 OK\r\nContent-Length: 5\r\nTransfer-Encoding: chunked\r\n\r\n5\r\nhello\r\n0\r\n\r\n"

	got, err := nodes.ParseResponse(ctx, ax, &gen.ParseResponseInput{Data: []byte(raw)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for ambiguous framing, got %+v", got)
	}
}

func TestParseResponse_Malformed(t *testing.T) {
	ctx := context.Background()
	ax := newTestContext(t)

	got, err := nodes.ParseResponse(ctx, ax, &gen.ParseResponseInput{Data: []byte("NOT A RESPONSE\r\n\r\n")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Ok {
		t.Fatalf("expected ok=false for malformed input, got %+v", got)
	}
}
