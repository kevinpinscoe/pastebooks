# Paste Books

Self-hosted webapp to store past buffers in different
books which can be shared with other users. On each book
is a paste buffers stylized as charms (shapes and colors)
for easy memory. The idea is for frequent paste buffers
for yourself or a team.

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
```yaml
port: 8080
jwt_secret: "change-me-super-secret"
database:
dsn: "youruser:yourpass@tcp(localhost:3306)/charmsdb?parseTime=true&charset=utf8mb4"
```
Environment variables override YAML:
- `PORT`
- `JWT_SECRET`
- `DB_DSN`


## Shapes & Colors
Shapes: `square, star, circle, triangle, rectangle, diamond, heart, clover, spade, hexagon, squiggle`


Colors: `red, green, blue, yellow, purple, pink, gold, black, orange, darkgray`


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