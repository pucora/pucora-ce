# SOAP Integration

**Edition:** Community Edition (Pucora CE)  
**Namespace:** `backend/soap`  
**Official docs:** [Pucora SOAP Backend](../../pucora-ce/docs/soap.md) · [KrakenD SOAP](https://www.krakend.io/docs/enterprise/backends/soap/)

## What it does

Connects KrakenD to legacy SOAP/XML services and exposes them as modern REST APIs. Clients send JSON (or GET requests) while KrakenD crafts SOAP XML requests using Go templates and transforms responses back to JSON, XML, or other encodings.

## Key capabilities

- Template-driven SOAP body construction with dynamic variables
- Transparent XML ↔ JSON encoding conversion
- Expose SOAP `POST` backends as `GET` REST endpoints for end users
- Response manipulation (`target`, `mapping`, `deny`) on XML responses

## How it works

```
Client GET /country/US
    │
    ▼
KrakenD loads SOAP template → replaces {{ .req_params.Country }} with "US"
    │
    ▼
POST XML to SOAP server (webservices.example.com)
    │
    ▼
XML response → traverse with target → map fields → return JSON to client
```

1. Endpoint receives REST request (any method exposed to client).
2. `backend/soap` compiles a Go text template into the SOAP XML body.
3. KrakenD POSTs XML to the SOAP `host` + `url_pattern`.
4. Response is decoded (`encoding: xml`) and optionally transformed before returning.

## Configuration

### Minimum setup

```json
{
  "endpoint": "/country/{country}",
  "method": "GET",
  "backend": [{
    "host": ["http://webservices.oorsprong.org"],
    "url_pattern": "/websamples.countryinfo/CountryInfoService.wso",
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

### `backend/soap` fields

| Field | Description |
|-------|-------------|
| `path` | Path to external Go template file (`.xml`) |
| `template` | Inline base64-encoded template (alternative to `path`) |
| `content_type` | Content-Type sent to SOAP server (default: `text/xml`) |
| `debug` | Log template variables and generated body (dev only) |

### Template variables

| Variable | Source |
|----------|--------|
| `.req_body` | Request body (requires `input_headers` with Content-Type) |
| `.req_params` | URL `{placeholders}` — first letter capitalized |
| `.req_headers` | Headers listed in `input_headers` |
| `.req_querystring` | Query strings listed in `input_query_strings` |
| `.req_path` | Backend `url_pattern` |

Supported body Content-Types for `.req_body`: `application/json`, `application/xml`, `text/xml`, `application/x-www-form-urlencoded`, `multipart/form-data`, `text/plain`.

### Example template (`soap_flag_request.xml`)

```xml
<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
    <CountryFlag xmlns="http://www.oorsprong.org/websamples.countryinfo">
      <sCountryISOCode>{{ .req_params.Country }}</sCountryISOCode>
    </CountryFlag>
  </soap:Body>
</soap:Envelope>
```

## Usage tips

- Set `output_encoding` on the endpoint to control client response format (default: `json`).
- Use `debug: true` during development to see compiled templates in logs.
- Combine with `target`, `mapping`, and `deny` to shape XML into clean JSON.
