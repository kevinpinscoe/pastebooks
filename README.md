# Paste Books

Self-hosted webapp to store paste buffers in different
books which can be shared with other users. On each book
is a paste buffers stylized as charms (shapes and colors)
for easy memory. The idea is for frequent paste buffers
for yourself or a team on a private network.

It is not bullet proof for public access so USE AT YOUR OWN RISK.

## Quick start (local dev)
```bash
cp config.example.yaml config.yaml
# Edit DB settings or use docker-compose

# Env is passed sets auth disabled mode
`AUTH_DISABLED=1 (or true)`

# Provision DB (via compose)
docker compose up -d db
# Create schema
docker exec -i pastebooks-db mysql -uroot -prootpass charmsdb < schema.sql

# Run backend + frontend
make run
# Open http://localhost:8080
```

## Production
```bash
docker build -t ghcr.io/kevinpinscoe/pastebooks:dev .
docker run --rm -p 8080:8080 \
-v $(pwd)/config.yaml:/app/config.yaml \
ghcr.io/kevinpinscoe/pastebooks:dev
```

## Environment
- Go 1.22+
- MySQL 8.x (or MariaDB 10.6+)

## Configuration (`config.yaml`)

`cookie_secure: true # The cookie will include the Secure flag.`

Browsers will only send it over HTTPS.

If you're running pastebooks behind TLS (e.g. via Apache reverse proxy or Nginx), this is recommended. If you try to use it over plain http://localhost:8080, the browser will not send the cookie — logins won't "stick".

`cookie_secure: false # The cookie will be sent over both HTTP and HTTPS. Use this only for local testing (e.g. localhost:8080 without HTTPS)`

Use `cookie_secure` in cooperation with `auth_disabled` for local development.

```yaml
port: 8080
jwt_secret: "change-me-super-secret"
auth_disabled: false         # <— add this (true in dev; false in prod)
cookie_secure: true
database:
dsn: "youruser:yourpass@tcp(localhost:3306)/charmsdb?parseTime=true&charset=utf8mb4"
```
Environment variables override YAML:
- `PORT`
- `JWT_SECRET`
- `DB_DSN`

## Authentication

See [Authentication](./auth.md).

## Shapes & Colors

Shapes: 

square, star, circle, triangle, rectangle, diamond, heart, clover, spade, hexagon, squiggle

Colors: red, green, blue, yellow, purple, pink, gold, black, orange, darkgray.

See [Best Color and Shape Pairings for Memorability](./memorability.md).


## API (summary)

- `POST /api/register {email, passcode}`
- `POST /api/login {email, passcode}` → sets `auth` HttpOnly cookie
- `POST /api/logout`
- `GET /api/me`
- Books (auth required):
- `GET /api/books` (mine)
- `POST /api/books` {title, note, is_public}
- `GET /api/books/:id` (owner)
- `PUT /api/books/:id` {title?, note?, is_public?}
- `DELETE /api/books/:id`
- Public read:
- `GET /api/public/books/:id` (no auth, returns read-only)
- Charms (owner):
- `GET /api/books/:id/charms`
- `POST /api/books/:id/charms` {shape, color, title, text_value}
- `PUT /api/charms/:id` {shape?, color?, title?, text_value?}
- `DELETE /api/charms/:id`


## Production notes
- Use a strong `JWT_SECRET`, set secure cookies, and serve via TLS/HTTPS behind a reverse proxy.
- Consider rate-limiting `/api/register` and `/api/login`.
- Add CSRF protection if adding state-changing endpoints consumed by browsers across origins.

## Proxies

See the [Proxy setup guide](./proxy.md).

## Database

See the [Database management guide](./database.md).

## Supply chain — verifying a published image

Every image published to `ghcr.io/kevinpinscoe/pastebooks` carries three attestations
bound to its digest: an **SPDX SBOM** (what is inside), **SLSA provenance** (where and
how it was built), and a **Cosign signature** (who published it, and whether it has
changed since).

```bash
IMG=ghcr.io/kevinpinscoe/pastebooks:latest

# What is inside the image
docker buildx imagetools inspect "$IMG" --format '{{ json .SBOM }}'

# Where and how it was built
docker buildx imagetools inspect "$IMG" --format '{{ json .Provenance }}'

# Who published it (tighten both regexes for real verification)
cosign verify "$IMG" \
  --certificate-identity-regexp='.*' \
  --certificate-oidc-issuer-regexp='.*'

# Known CVEs in the published image
grype "$IMG"
```

SPDX documents for each platform are also attached to the GitHub release. Those are a
convenience copy for reading without a registry client — **the registry attestation is
the source of truth**, and the two can drift.

### What the SBOM does not tell you

An SBOM is an inventory, not a clean bill of health. This image can have a complete,
correctly signed SBOM and still contain known CVEs — the SBOM is what makes finding
them possible. Specific limits that apply to *this* image:

- **The final stage is distroless and receives a statically linked Go binary.** A
  scan of only that stage would list the distroless base packages and no Go modules
  at all — an inventory that looks complete and is not. The `backend` and `fe` stages
  are therefore marked `BUILDKIT_SBOM_SCAN_STAGE=true` in the `Dockerfile` so their
  dependencies reach the attestation. If you add a build stage whose dependencies do
  not survive into the final image, mark it too.
- **The frontend is copied in as built assets**, not as installed packages, so
  anything it vendors is only as visible as the `fe` stage scan makes it.
- **`config.example.yaml` is `COPY`'d in** without package metadata and will not
  appear as a component.
- **Being listed is not being reachable.** A package can appear with a CVE whose
  vulnerable code this image never calls. That is what VEX is for; there is no
  `.vex/openvex.json` here yet because nothing is being blocked (see below).
- **Generators disagree.** Two SBOMs of this image, from different tools, will not
  match line for line.

### The CVE gate reports, it does not block

The release workflow scans the pushed digest with Grype at `severity-cutoff: high`,
uploads the result to the repository's **Security** tab, and **does not fail the
build**. That is deliberate, not an oversight.

Measured 2026-08-04, before the gate was added, this image carried **29 findings at
or above high** (3 critical, 26 high), essentially all inherited from the Debian base
of the distroless image. Turning on `fail-build: true` would have meant no release
could ever publish until every one was fixed or waived.

When that baseline has been triaged, flip `fail-build` to `true` in
`.github/workflows/build-and-push.yml`. Two things not to do instead: do not raise
`severity-cutoff` to make findings disappear, and do not delete the step. A CVE that
genuinely does not affect this image is dispositioned with an OpenVEX statement at
`.vex/openvex.json`, committed and reviewable.

**Known gap:** Grype resolves a multi-arch manifest to the runner's own platform, so
the gate covers `linux/amd64` only. The `linux/arm64` image is built, signed, and has
its own SBOM, but nothing scans it for CVEs.

**Rebuild trigger:** a digest's CVE posture is frozen at build time and only degrades.
Cut a new release when the base image or a dependency updates — a scan that was clean
six months ago says nothing about today.
