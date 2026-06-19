#!/usr/bin/env bash
# Starts postgres, backend, and frontend. Press Ctrl+C to stop everything.

set -e

# Load backend env (gitignored) if present, then fall back to dev defaults.
set -a
[ -f backend/configs/.env ] && . backend/configs/.env
set +a
: "${DATABASE_URL:=postgres://gochess:gochess@localhost:5432/gochess}"
: "${JWT_SECRET:=dev-secret-not-for-production-0123456789abcd}"
: "${COOKIE_SECURE:=false}"
export DATABASE_URL JWT_SECRET COOKIE_SECURE

cleanup() {
  echo ""
  echo "[dev] shutting down..."
  kill 0
  docker compose down
}
trap cleanup EXIT

echo "[dev] starting postgres..."
docker compose up -d

echo "[dev] waiting for postgres..."
until docker compose exec -T postgres pg_isready -U gochess >/dev/null 2>&1; do
  sleep 1
done

if [ ! -d frontend/node_modules ]; then
  echo "[dev] installing frontend dependencies (first run)..."
  (cd frontend && npm install)
fi

echo "[dev] starting backend on :8080"
(cd backend && go run ./cmd/server) &

echo "[dev] starting frontend on :5173"
(cd frontend && npm run dev) &

echo "[dev] ready. open http://localhost:5173  (Ctrl+C to stop)"
wait
