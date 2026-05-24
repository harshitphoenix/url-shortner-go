# url-shortner-go

A lightweight URL shortener HTTP service written in Go using only the standard library.

## Features

- Shorten any `http`/`https` URL to a cryptographically random base-62 code
- Redirect short codes to their original URLs (301)
- Per-link click stats
- Structured JSON logging (`log/slog`)
- Middleware: request logger, 1 MB body limit, JSON content-type enforcement
- Thread-safe in-memory store (swap-friendly via `domain.Repository` interface)

## Project Layout

```
cmd/server/          entry point
config/              env-based configuration
internal/
  domain/            core entities & interfaces (URL, Repository, Service)
  handler/           HTTP handlers + middleware
  repository/memory/ in-memory Repository implementation
  service/           business logic (Shorten, Resolve, Stats)
pkg/generator/       crypto/rand base-62 code generator
```

## API

| Method | Path             | Description                        |
|--------|------------------|------------------------------------|
| POST   | `/shorten`       | Create a short URL                 |
| GET    | `/{code}`        | Redirect to the original URL (301) |
| GET    | `/stats/{code}`  | Click stats for a short code       |
| GET    | `/health`        | Liveness probe                     |

### POST /shorten

Request (`Content-Type: application/json` required):

```json
{ "url": "https://example.com/some/long/path" }
```

Response `201 Created`:

```json
{
  "code": "aB3xY7z",
  "short_url": "http://localhost:8080/aB3xY7z",
  "original_url": "https://example.com/some/long/path"
}
```

### GET /stats/{code}

Response `200 OK`:

```json
{
  "code": "aB3xY7z",
  "original_url": "https://example.com/some/long/path",
  "short_url": "http://localhost:8080/aB3xY7z",
  "clicks": 42,
  "created_at": "2026-05-24T10:00:00Z"
}
```

## Configuration

All settings are read from environment variables.

| Variable      | Default                      | Description                              |
|---------------|------------------------------|------------------------------------------|
| `PORT`        | `8080`                       | Port the server listens on               |
| `BASE_URL`    | `http://localhost:<PORT>`    | Public base URL prepended to short codes |
| `CODE_LENGTH` | `7`                          | Short code length (clamped to 4–20)      |

## Running

```bash
go run ./cmd/server
```

With custom settings:

```bash
PORT=9090 BASE_URL=https://go.sh CODE_LENGTH=8 go run ./cmd/server
```

## Example

```bash
# shorten a URL
curl -s -X POST http://localhost:8080/shorten \
  -H 'Content-Type: application/json' \
  -d '{"url": "https://go.dev/doc/effective_go"}' | jq

# follow the redirect
curl -L http://localhost:8080/aB3xY7z

# check stats
curl http://localhost:8080/stats/aB3xY7z | jq
```

## Notes

- Storage is in-memory only — all data is lost on restart. Implement `domain.Repository` to plug in a persistent backend (e.g. Redis, Postgres).
- Short codes are generated with `crypto/rand`, making them unpredictable.
