# AGENTS.md

Guidance for AI coding agents working in this repository. This is the main instructions file;
`CLAUDE.md` just imports it. For which AI model to use for which kind of task, see [MODELS.md](MODELS.md).
Personal or machine-specific notes go in a gitignored `CLAUDE.local.md` (Claude Code loads it
automatically) or in an untracked `.internal/` folder. Never copy their contents into tracked files.

<!-- The next line makes Claude Code load MODELS.md automatically; other tools can ignore it. -->
@MODELS.md

## Project

Capture Maps stitches many Google Maps Static API tiles into one high-resolution PNG.
It is a monorepo:

- **Backend (repo root)**: Go HTTP API, module `github.com/zxcv32/capture-maps`.
- **Frontend (`ui/`)**: React 18 (Create React App) using `@react-google-maps/api` and react-bootstrap.

## Layout

| Path | Purpose |
|---|---|
| `cmd/capturemapsapi/` | `main`: loads config with viper, registers `/` (redirect) and `/print` |
| `internal/service/print.go` | `/print` handler: decode, validate, fetch tile rows, merge, stream PNG |
| `internal/service/ratelimit.go` | Per-client-IP token-bucket rate limiter |
| `pkg/earth/` | Lat/lng to tile math (`projections.go`) and Static API fetch (`compute.go`) |
| `pkg/file/` | Image download, temp filenames, and merging grids with `go-image-merge` |
| `test/earth_test/` | Tests for the tile math |
| `api/request-schema.json` | JSON Schema for the `/print` request body |
| `configs/config.toml.template` | Config template; copy to `configs/config.toml` (gitignored) |
| `deploy/` | Docker Compose files for dev and prod |
| `ui/src/App.js` | Form and the call to `${REACT_APP_API_HOST}/print` |
| `ui/src/component/Map.js` | Google Map picker (center, zoom, map type) |

## Commands

Backend (from repo root):

```bash
cp configs/config.toml.template configs/config.toml   # then set the api_key
go run ./cmd/capturemapsapi                            # serves on :8090
go build ./...
go test ./...
go vet ./...
```

`GOOGLE_MAPS_API_KEY` in the environment overrides `google-maps.api_key` in the config.

Frontend (from `ui/`):

```bash
cp .env.template .env   # set REACT_APP_GOOGLE_MAPS_API_KEY and REACT_APP_API_HOST
npm ci
npm start
npm test
npm run build
```

Docker: `cd deploy && docker compose up --build -d`.

## The `/print` contract

`POST /print` with JSON `{lat, lng, zoom, radius, mapTypeId}` returns `image/png`.

- `lat` in [-90, 90], `lng` in [-180, 180], `zoom` in [0, 21]
- `radius` in [1, `maxRadius`] (currently 5). `radius` counts tiles out from the centre tile, inclusive.
- `mapTypeId` is one of `hybrid`, `satellite`, `roadmap`, `terrain`

If you change any of these rules, update `validateRequest` in `print.go`, `api/request-schema.json`,
the README, and the UI together.

## Rules

- **Every tile is a billable Google API call.** A request makes up to `(2*radius-1)^2` calls.
  Do not raise `maxRadius`, loosen the rate limit, or add retry loops without saying so explicitly.
- **Never commit API keys.** `configs/config.toml` and `ui/.env` are gitignored; keep it that way.
  Use the templates for new settings.
- **Clean up temp files.** Tiles and row images are written to disk. Every error path in
  `captureTiles` / `westToEast` must delete what it created, and so must new code.
- **Keep the Google watermark.** The Google Maps Platform Terms forbid removing branding.
  Don't crop or mask it.
- Return errors instead of panicking or calling `log.Fatal` outside `main`.
  Client mistakes get 400; internal failures get 500 with a generic message.
- Go code must pass `gofmt` and `go vet`. Add or update `_test.go` files next to the code you change.
- Keep production dependencies minimal. Pin vulnerable transitive npm packages via `overrides` in `ui/package.json`.

## Known gotchas

- Go versions are inconsistent: `go.mod` says 1.19, CI (`.github/workflows/go.yml`) uses 1.18,
  and the `Dockerfile` builds with 1.22. `print.go` defines its own `min` helper, so the code
  still builds on all three.
- The `radius` range in `api/request-schema.json` ([1,15]) is out of date; the code caps it at 5.
- The limiter map in `ratelimit.go` never evicts entries (see the comment there).

## Commits

Use Conventional Commit prefixes (`fix:`, `feat:`, `chore:`, ...) as in the existing history.
