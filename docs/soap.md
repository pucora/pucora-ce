# SOAP Backend Integration

Velonetics connects legacy SOAP/XML services to modern REST clients. The gateway crafts SOAP request bodies from Go templates, optionally applies WS-Security, and transforms XML responses to JSON (or other encodings).

Implemented by [`velonetics-soap`](https://github.com/velonetics/velonetics-soap) via `extra_config.backend/soap` on a backend.

## Quick start

See the Docker example in [`examples/soap/`](../examples/soap/):

```bash
cd velonetics-ce-master
make soap-compose-test
```

Minimal config:

```json
{
  "endpoint": "/country/{country}",
  "method": "GET",
  "backend": [{
    "host": ["http://soap-server:8081"],
    "url_pattern": "/CountryInfoService.wso",
    "method": "POST",
    "encoding": "xml",
    "extra_config": {
      "backend/soap": {
        "path": "./soap_flag_request.xml"
      }
    },
    "target": "Envelope.Body.CountryFlagResponse",
    "mapping": { "CountryFlagResult": "flag_url" },
    "deny": ["-m"]
  }]
}
```

## Template variables

| Variable | Source |
|----------|--------|
| `.req_params` | URL `{placeholders}` (first letter capitalized) |
| `.req_headers` | Endpoint `input_headers` |
| `.req_querystring` | Endpoint `input_query_strings` |
| `.req_path` | Backend `url_pattern` |
| `.req_body` | Client body (requires `Content-Type` in `input_headers`) |

## Configuration reference

### Core fields

| Field | Description |
|-------|-------------|
| `path` | File path to Go template (XML body) |
| `template` | Base64-encoded inline template |
| `content_type` | Sent to SOAP server (default: `text/xml`) |
| `debug` | Log template variables and generated body |

### SOAPAction

| Field | Description |
|-------|-------------|
| `soap_action` | Sets HTTP `SOAPAction` header (SOAP 1.1, quoted automatically) |

If omitted, Velonetics can derive the action from WSDL when `wsdl` is configured.

### WSDL (parse-only)

| Field | Description |
|-------|-------------|
| `wsdl.path` / `wsdl.url` | Load WSDL from file or HTTP |
| `wsdl.service` | Service name (optional if only one) |
| `wsdl.port` | Port name (optional if only one) |
| `wsdl.operation` | Operation name for SOAPAction lookup |
| `wsdl.generate_template` | Generate minimal envelope when no `path`/`template` |

WSDL parsing does not modify your config file; it supplies runtime hints (SOAPAction, location) and optional starter templates.

### Template hot-reload

| Field | Description |
|-------|-------------|
| `watch_template` | Reload `path` templates on file change (`fsnotify`) |
| `reload_interval` | Poll interval (e.g. `5s`) for `path` templates |

Inline `template` values are loaded once at startup.

### WS-Security

| Field | Description |
|-------|-------------|
| `ws_security.username_token.username` | WS-Security username |
| `ws_security.username_token.password` | Password (supports `{{ env "VAR" }}`) |
| `ws_security.saml_assertion.path` | Static SAML assertion XML file |
| `ws_security.x509.cert_path` | Client certificate PEM |
| `ws_security.x509.key_path` | Private key PEM |
| `ws_security.x509.key_password` | Optional key password |

WS-Security is applied after template rendering. The body must be a SOAP `Envelope`.

## Response shaping

Use standard backend fields:

- `encoding: "xml"` on the backend
- `target`, `mapping`, `deny` to shape JSON for clients
- `output_encoding` on the endpoint (default: `json`)

## Combining with other middleware

- **Martian** (`modifier/martian`) for extra HTTP headers not covered by SOAPAction
- **JWT** (`auth/validator`) on the endpoint for client authentication

## See also

- [Non-REST connectivity reference](../../docs/non-rest-connectivity/01-soap.md)
- KrakenD SOAP docs (baseline parity): https://www.krakend.io/docs/enterprise/backends/soap/
