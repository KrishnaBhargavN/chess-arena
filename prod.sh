#!/usr/bin/env bash
# Production launcher: Postgres + Go backend + Caddy (serving the built frontend).
# Expose it with cloudflared, run separately (see deploy/cloudflared.yml.example).
#
# Requires backend/configs/.env with real prod values:
#   JWT_SECRET     = output of: openssl rand -base64 48
#   COOKIE_SECURE  = true
#   ALLOWED_ORIGIN = https://play.yourdomain.com
#   DATABASE_URL   = postgres://gochess:<password>@localhost:5432/gochess
# and a root .env with POSTGRES_PASSWORD matching DATABASE_URL.
set -e

set -a
. backend/configs/.env
set +a
: "${JWT_SECRET:?set JWT_SECRET in backend/configs/.env}"
: "${ALLOWED_ORIGIN:?set ALLOWED_ORIGIN in backend/configs/.env}"

cleanup() {
  echo ""
  echo "[prod] stopping backend + caddy (postgres left running)..."
  kill 0
}
trap cleanup EXIT

echo "[prod] starting postgres..."
docker compose up -d
until docker compose exec -T postgres pg_isready -U "${POSTGRES_USER:-gochess}" >/dev/null 2>&1; do
  sleep 1
done

echo "[prod] building frontend..."
(cd frontend && npm install && npm run build)

echo "[prod] building backend..."
(cd backend && go build -o ../bin/chess-server ./cmd/server)

echo "[prod] starting backend on :8080"
bin/chess-server &

echo "[prod] starting caddy on 127.0.0.1:8081"
caddy run --config Caddyfile &

echo "[prod] up — point cloudflared at http://localhost:8081  (Ctrl+C to stop app)"
wait
