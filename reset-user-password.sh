#!/usr/bin/env bash

set -euo pipefail

# --- Config (override via env if needed) ---
CONTAINER="${CONTAINER:-pastebooks-db}"
DB_NAME="${DB_NAME:-charmsdb}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PASS="${MYSQL_PASS:-rootpass}"
BCRYPT_ROUNDS="${BCRYPT_ROUNDS:-12}"   # 10-14 is common

# --- Dependencies ---
need() { command -v "$1" >/dev/null 2>&1 || { echo "Missing required tool: $1" >&2; exit 1; }; }
need docker
need htpasswd

# --- Email to update (arg or prompt) ---
EMAIL="${1:-}"
if [[ -z "$EMAIL" ]]; then
  read -rp "Email to reset: " EMAIL
fi
if [[ -z "$EMAIL" ]]; then
  echo "No email provided." >&2; exit 1
fi

# --- Prompt for new password (hidden) ---
while :; do
  read -srp "New password: " NEWPASS; echo
  read -srp "Confirm password: " CONFIRM; echo
  if [[ "$NEWPASS" != "$CONFIRM" ]]; then
    echo "Passwords do not match. Try again." >&2
    continue
  fi
  if [[ -z "$NEWPASS" ]]; then
    echo "Password cannot be empty." >&2
    continue
  fi
  break
done

# --- Generate bcrypt hash ($2y$...) ---
HASH="$(htpasswd -bnBC "$BCRYPT_ROUNDS" "" "$NEWPASS" | tr -d ':\n')"
# Escape single quotes for SQL
SQL_HASH="${HASH//\'/\'\'}"

# --- Sanity: ensure container is running ---
if ! docker ps --format '{{.Names}}' | grep -qx "$CONTAINER"; then
  echo "Container '$CONTAINER' is not running." >&2
  exit 1
fi

# --- Update pass_hash (stored as VARBINARY) ---
SQL="
SELECT COUNT(*) AS user_exists FROM users WHERE email='${EMAIL}';
UPDATE users SET pass_hash=_binary '${SQL_HASH}' WHERE email='${EMAIL}';
SELECT ROW_COUNT() AS changed_rows;
"

# Use MYSQL_PWD so the password isn't visible in process args
docker exec -e MYSQL_PWD="$MYSQL_PASS" -i "$CONTAINER" \
  mysql -u"$MYSQL_USER" -D "$DB_NAME" -e "$SQL"

echo "Password updated for ${EMAIL}.
Note: active sessions/JWTs may still work until your app invalidates them."
