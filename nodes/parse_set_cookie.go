package nodes

import (
	"context"
	"net/http"

	"christiangeorgelucas/http-message-tools/axiom"
	gen "christiangeorgelucas/http-message-tools/gen"
)

// Parse the raw value of a single response Set-Cookie header into its
// structured attributes (name, value, Domain, Path, Expires, Max-Age,
// Secure, HttpOnly, SameSite), via net/http.ParseSetCookie. Call once per
// Set-Cookie occurrence when a response has several (use ParseResponse or
// GetHeader to obtain each occurrence first).
func ParseSetCookie(ctx context.Context, ax axiom.Context, input *gen.ParseSetCookieInput) (*gen.ParseSetCookieOutput, error) {
	if input == nil || input.Value == "" {
		return &gen.ParseSetCookieOutput{Error: "value is required"}, nil
	}
	if len(input.Value) > maxLineBytes {
		return &gen.ParseSetCookieOutput{Error: "value exceeds the 8192 byte limit"}, nil
	}

	c, err := http.ParseSetCookie(input.Value)
	if err != nil {
		return &gen.ParseSetCookieOutput{Error: "parse error: " + err.Error()}, nil
	}

	out := &gen.ParseSetCookieOutput{
		Ok:        true,
		Name:      c.Name,
		Value:     c.Value,
		Domain:    c.Domain,
		Path:      c.Path,
		MaxAge:    int64(c.MaxAge),
		HasMaxAge: c.MaxAge != 0,
		Secure:    c.Secure,
		HttpOnly:  c.HttpOnly,
	}
	if !c.Expires.IsZero() {
		// http.TimeFormat ("Mon, 02 Jan 2006 15:04:05 GMT") is the correct
		// choice here, not time.RFC1123: RFC1123's zone token renders the
		// *name* of the time.UTC Location, which Go prints as "UTC", not
		// the "GMT" suffix HTTP dates require (RFC 7231 §7.1.1.1).
		out.Expires = c.Expires.UTC().Format(http.TimeFormat)
	}
	switch c.SameSite {
	case http.SameSiteStrictMode:
		out.SameSite = "Strict"
	case http.SameSiteLaxMode:
		out.SameSite = "Lax"
	case http.SameSiteNoneMode:
		out.SameSite = "None"
	}
	return out, nil
}
