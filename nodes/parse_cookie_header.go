package nodes

import (
	"context"
	"net/http"

	"christiangeorgelucas/http-message-tools/axiom"
	gen "christiangeorgelucas/http-message-tools/gen"
)

// Parse the raw value of a request's Cookie header (e.g. "a=1; b=2") into
// name/value pairs, via net/http.Request.Cookies() — the same RFC 6265
// cookie-string parser net/http itself uses to expose r.Cookies() on an
// incoming server request.
func ParseCookieHeader(ctx context.Context, ax axiom.Context, input *gen.ParseCookieHeaderInput) (*gen.ParseCookieHeaderOutput, error) {
	if input == nil || input.Value == "" {
		return &gen.ParseCookieHeaderOutput{Error: "value is required"}, nil
	}
	req := &http.Request{Header: http.Header{"Cookie": []string{input.Value}}}
	cookies := req.Cookies()

	out := &gen.ParseCookieHeaderOutput{Ok: true}
	for _, c := range cookies {
		out.Cookies = append(out.Cookies, &gen.CookiePair{Name: c.Name, Value: c.Value})
	}
	return out, nil
}
