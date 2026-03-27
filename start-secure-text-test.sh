#!/bin/bash
set -e

BRANCH=$(git rev-parse --abbrev-ref HEAD)
PROJECT="pastebooks-${BRANCH//\//-}"
IMAGE="${PROJECT}:test"
PORT=$(free-port)

OVERRIDE="/tmp/${PROJECT}-override.yml"

teardown() {
  echo ""
  echo "Tearing down $PROJECT..."
  docker compose -p "$PROJECT" \
    -f docker-compose.yml \
    -f "$OVERRIDE" \
    down -v
  docker rmi "$IMAGE" 2>/dev/null || true
  rm -f "$OVERRIDE"
  echo "Done."
}

cat > "$OVERRIDE" <<EOF
services:
  app:
    image: ${IMAGE}
    build: .
    ports:
      - "${PORT}:8080"
    container_name: ${PROJECT}-app
    environment:
      AUTH_DISABLED: "1"
      COOKIE_SECURE: "false"
  db:
    container_name: ${PROJECT}-db
EOF

echo "==> Building image: $IMAGE"
docker build -t "$IMAGE" .

echo "==> Starting DB..."
docker compose -p "$PROJECT" -f docker-compose.yml -f "$OVERRIDE" up -d db

echo "==> Waiting for DB to be healthy..."
until [ "$(docker inspect ${PROJECT}-db --format '{{.State.Health.Status}}' 2>/dev/null)" = "healthy" ]; do
  sleep 2
done

echo "==> Applying schema..."
docker exec -i "${PROJECT}-db" mysql -uroot -prootpass charmsdb < schema.sql 2>&1 | grep -v "password on the command line" || true

echo "==> Starting app on port $PORT..."
docker compose -p "$PROJECT" -f docker-compose.yml -f "$OVERRIDE" up -d app

echo ""
echo "  Test stack running at http://localhost:${PORT}"
echo "  Project: $PROJECT | Image: $IMAGE"
echo ""
echo "  Press Ctrl+C to stop and tear down."

trap teardown INT TERM
docker compose -p "$PROJECT" -f docker-compose.yml -f "$OVERRIDE" logs -f app
