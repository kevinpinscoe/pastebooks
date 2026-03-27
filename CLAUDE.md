# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

Pastebooks is a self-hosted paste buffer manager. Users organize pastes ("charms") into "books". Each charm has a shape + color combination for visual identification. The app is designed for private network deployment.

## Commands

```bash
make run      # Run backend dev server (Go, serves frontend at localhost:8080)
make build    # Build Docker image (pastebooks:dev)
make fmt      # Format Go code
make lint     # Run go vet
```

There are no automated tests. Manual scripts exist: `test_login.sh`, `reset-user-password.sh`, `rebuild-compose.sh`.

**Finding a free port:** Use `free-port` (`~/bin/free-port`) — picks a random available port in the 20000–40000 range. Used by feature branch test scripts to avoid collisions with the dev stack.

**Local dev with Docker:**
```bash
docker compose up -d db          # Start MySQL
make run                         # Run backend (reads config.yaml)
```

**Full Docker stack:**
```bash
make build
docker compose up -d
```

## Architecture

**Stack:** Go 1.25 backend (Gin) + Vanilla JS SPA (no build step) + MySQL 8.x

**Backend layout** (`backend/`):
- `main.go` — entry point, routing setup
- `auth.go` — HMAC-signed compact tokens, bcrypt password hashing, cookie management
- `config.go` — YAML config loading with environment variable overrides
- `handlers_books.go` / `handlers_charms.go` — REST API handlers
- `middleware.go` — auth middleware, user context injection
- `models.go` — User, Book, Charm structs
- `db.go` — database connection setup

**Frontend** (`frontend/`) — single `index.html` + `app.js` SPA, no framework, no build step. Served by Go under `/static/`.

**Routing pattern:**
- `/api/*` — REST API (Gin routes)
- `/static/*` — static assets
- All other GET/HEAD — SPA fallback to `index.html`

## Configuration

`config.yaml` is the primary config file (see `config.example.yaml` for template). Environment variables override YAML values:

| Env Var | YAML Key | Notes |
|---|---|---|
| `PORT` | `port` | Default 8080 |
| `JWT_SECRET` | `jwt_secret` | Required in prod |
| `DB_DSN` | `database.dsn` | MySQL connection string |
| `AUTH_DISABLED` | `auth_disabled` | `true`/`1` auto-creates dev-user |
| `COOKIE_SECURE` | `cookie_secure` | Set `false` for localhost dev |

For local dev without Docker DB, set `auth_disabled: true` and `cookie_secure: false` in `config.yaml`.

## Data Model

Three tables: `users` → `books` → `charms`. All PKs are UUIDs.

Charm `shape` and `color` are MySQL ENUMs — validated server-side. Valid shapes: `square star circle triangle rectangle diamond heart clover spade hexagon squiggle`. Valid colors: `red green blue yellow purple pink gold black orange darkgray`.

## Deployment

CI/CD (`.github/workflows/`) triggers on `v*.*.*` tags, builds multi-arch images (amd64/arm64), and pushes to GHCR. See `proxy.md` for reverse proxy setup and `database.md` for backup/migration procedures.
