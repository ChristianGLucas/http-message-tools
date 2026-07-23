# http-message-tools

Composable [Axiom](https://axiomide.com) nodes for deterministic parsing and building of raw
**HTTP/1.x request and response messages** in their wire text form — the application-layer
HTTP message format. Distinct from `packet-tools` (raw network/transport packet decoding),
`url-tools` (URL parsing), `mime-tools` (media-type headers), `useragent-tools` (the
User-Agent header only), and `email-tools` (RFC822/MIME email).

Wraps Go's standard library — `net/http`, `net/http/httputil`, `net/textproto` — with **zero
third-party dependency**: the same RFC 7230/9110 grammar, Content-Length/chunked transfer
decoding, and cookie parsing `net/http` itself uses in production, plus thin, order-preserving
glue where `net/http`'s own map-based `Header` type would otherwise silently discard wire
order, and where `net/http` silently resolves a Content-Length + Transfer-Encoding
combination rather than flagging it (a classic request-smuggling ambiguity this package
rejects explicitly).

Every node is a pure, deterministic single-input to single-output transform: no network
calls, no sockets, no wall-clock, no randomness. Input is bounded to 4 MiB (the Axiom node
transport ceiling), header count is capped, and chunked-body decoding output is capped
regardless of a declared chunk size.

## Use it from your agent or app

Every node in this package is a **live, auto-scaling API endpoint** on the
[Axiom](https://axiomide.com) marketplace — call it from an AI agent or your own
code, with nothing to self-host.

**📦 See it on the marketplace:**
https://dev.axiomide.com/marketplace/christiangeorgelucas/http-message-tools@0.1.0

**Hook it up to an AI agent (MCP).** Add Axiom's hosted MCP server to any MCP
client and every node becomes a typed tool your agent can call — search the
catalog, inspect a schema, and invoke it directly.

```bash
# Claude Code
claude mcp add --transport http axiom https://api.axiomide.com/mcp \
  --header "Authorization: Bearer $AXIOM_API_KEY"
```

Claude Desktop, Cursor, or any config-based client:

```json
{
  "mcpServers": {
    "axiom": {
      "type": "http",
      "url": "https://api.axiomide.com/mcp",
      "headers": { "Authorization": "Bearer YOUR_AXIOM_API_KEY" }
    }
  }
}
```

**Call it from the CLI.**

```bash
axiom invoke christiangeorgelucas/http-message-tools/ParseRequest --input '{ ... }'
```

**Call it over HTTP.**

```bash
curl -X POST https://api.axiomide.com/invocations/v1/nodes/christiangeorgelucas/http-message-tools/0.1.0/ParseRequest \
  -H "Authorization: Bearer $AXIOM_API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{ ... }'
```

> Input/output schema for each node is on the marketplace page above, or via
> `axiom inspect node christiangeorgelucas/http-message-tools/ParseRequest`.

### Get started free

Install the CLI:

```bash
# macOS / Linux — Homebrew
brew install axiomide/tap/axiom

# macOS / Linux — install script
curl -fsSL https://raw.githubusercontent.com/AxiomIDE/axiom-releases/main/install.sh | sh
```

**Windows:** download the `windows/amd64` `.zip` from the
[releases page](https://github.com/AxiomIDE/axiom-releases/releases), unzip it,
and put `axiom.exe` on your `PATH`.

Then `axiom version` to verify, `axiom login` (GitHub or Google) to authenticate,
and create an API key under **Console → API Keys**. Docs and sign-up at
**[axiomide.com](https://axiomide.com)**.

## Nodes

- **ParseRequest** — raw request bytes to method / target / version / headers / body.
- **ParseResponse** — raw response bytes to version / status code / reason / headers / body.
- **ParseMessage** — auto-detects request vs. response from the first line.
- **ParseRequestLine** — a bare request-line to method / target / version.
- **ParseStatusLine** — a bare status-line to version / status code / reason.
- **ParseHeaderBlock** — a raw header block to an order-preserving, repeat-preserving list.
- **GetHeader** — extract one named header's value(s), case-insensitively.
- **DechunkBody** — decode a `Transfer-Encoding: chunked` body into its raw bytes.
- **ParseCookieHeader** — a request `Cookie:` header value into name/value pairs.
- **ParseSetCookie** — a response `Set-Cookie:` header value into its structured attributes.
- **BuildRequest** — structured components to a raw request message (inverse of ParseRequest).
- **BuildResponse** — structured components to a raw response message (inverse of ParseResponse).
- **SplitRequestTarget** — a request-target into its path and raw query string.

## License

MIT — see [LICENSE](./LICENSE).

Built for the Axiom marketplace.
