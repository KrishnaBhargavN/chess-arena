# chess-arena

A real-time multiplayer chess server written in Go, with a React + TypeScript frontend. Two players log in, get matched, and play a live game over WebSockets — with persistence, replay, reconnect, and game-over detection.

## Features

- **Authentication** — register / login with HttpOnly JWT cookies, bcrypt-hashed passwords
- **Matchmaking** — FIFO queue, random color assignment, two-phase color delivery (HTTP + WS) to avoid a timing race
- **Live games** — moves stream over WebSocket between paired players in real time
- **Promotion picker** — choose Q / R / B / N on pawn promotion (no forced auto-queen)
- **Resign** — broadcast `game_over` to both players, mark game `finished` in the DB
- **Game-over detection** — checkmate, stalemate, draws, and resignation all detected via `notnil/chess`
- **Game replay** — step through any past game move-by-move with prev / next / start / end controls and keyboard shortcuts
- **Game history** — `My Games` lists every game you've played with opponent username + result
- **WebSocket reconnect** — exponential-backoff client wrapper, HTTP resync via `GET /games/:id` to recover state
- **Server-restart resilience** — game sessions lazily rebuilt from the DB; in-progress games survive a backend restart
- **PostgreSQL persistence** — users, games, and moves persisted; falls back to in-memory stores when `DATABASE_URL` is unset

## Tech stack

| Layer | Tools |
|---|---|
| Backend | Go, `gorilla/websocket`, `notnil/chess`, `golang-jwt/jwt/v5`, `jackc/pgx/v5`, `golang.org/x/crypto/bcrypt` |
| Frontend | React, TypeScript, Vite, `chess.js`, `chessground`, Axios, React Router |
| Infrastructure | PostgreSQL 16 (Docker Compose) |

## Quick start

```bash
./dev.sh
```

Starts Postgres in Docker, waits for it to be ready, then runs the backend (with `DATABASE_URL` set) and the Vite dev server. Press `Ctrl+C` to tear everything down including the container.

Open [http://localhost:5173](http://localhost:5173). To test multiplayer, open the site in two different browsers (or one regular + one incognito) so the auth cookies don't conflict.

### Requirements

- Go 1.22+
- Node.js 20+
- Docker (for Postgres)

### Manual setup

If you'd rather run the pieces separately:

```bash
# Postgres
docker compose up -d

# Backend (with DB)
cd backend
DATABASE_URL=postgres://gochess:gochess@localhost:5432/gochess go run ./cmd/server

# Backend (without DB — in-memory fallback, no persistence)
cd backend
go run ./cmd/server

# Frontend
cd frontend
npm install
npm run dev
```

> The Go server reads source at startup — file edits don't hot-reload. After backend changes, `Ctrl+C` and re-run. For auto-rebuild see [`air`](https://github.com/cosmtrek/air).

## Architecture at a glance

**Two WebSocket connection types per player**, multiplexed over the same `/ws` endpoint:

| Connection | Hub key | Auth payload | Purpose |
|---|---|---|---|
| Lobby WS | `playerID` | `{}` | Receives `match_found` |
| Game WS | `playerID:gameID` | `{ gameId }` | Sends and receives moves for one specific game |

`playerID` is always read server-side from the verified JWT cookie — never trusted from client payloads.

**Hybrid store**: active game chess state lives in memory for fast move validation; player records, game records, and move history are persisted to Postgres. On server restart, the in-memory chess state is lazily rebuilt by replaying the moves table.

**WebSocket reconnect**: a small `ReconnectingWebSocket` wrapper on the client handles backoff (500ms → 30s cap). After reconnect, the game page calls `GET /games/:id` and replays any missed moves — idempotent because `chess.js` silently rejects invalid moves.

## HTTP endpoints

| Method | Path | Auth | Purpose |
|---|---|---|---|
| POST | `/auth/register` | — | Create account |
| POST | `/auth/login` | — | Set JWT cookie |
| POST | `/auth/logout` | — | Clear cookie |
| GET | `/auth/me` | cookie | Restore session on page load |
| POST | `/matchmaking/join` | cookie | Join the queue |
| GET | `/games` | cookie | List the user's past games |
| GET | `/games/:id` | — | Full game state (FEN, moves, players) |
| GET | `/games/:id/moves` | cookie | Ordered move list |
| POST | `/games/:id/move` | cookie | Apply a move (turn-validated) |
| WS | `/ws` | cookie | Game / lobby WebSocket — cookie verified before upgrade |

## Project structure

```
go-chess/
├── docker-compose.yml          # Postgres for local dev
├── dev.sh                      # one-command dev runner
├── backend/
│   ├── cmd/server/main.go      # wires everything; reads DATABASE_URL
│   └── internal/
│       ├── auth/               # JWT, bcrypt, middleware, register/login
│       ├── db/                 # pgxpool connect + embedded migrations
│       ├── game/               # chess session wrapper, game manager
│       ├── handlers/           # HTTP handlers
│       ├── matchmaking/        # FIFO queue
│       ├── models/             # shared types (Game, MoveRecord)
│       ├── store/              # Store interface, MemoryStore, PostgresStore
│       └── ws/                 # Hub, WS upgrade, message routing
└── frontend/
    └── src/
        ├── context/            # AuthContext (api client), WSContext (lobby WS)
        ├── hooks/useChessGame  # chess.js state hook
        ├── lib/reconnectingWebSocket  # exponential-backoff WS wrapper
        └── components/         # Play, ChessBoard, Replay, GamesList, etc.
```

## Testing two players locally

Open two different browsers (or one regular + one incognito). Two tabs in the same browser share the cookie, so the second login overwrites the first.

1. Register two accounts (one per browser)
2. Click **Play** in both windows — both join matchmaking
3. As soon as both are queued, a game is created and both browsers navigate to `/play/:gameId`
4. Move pieces, promote pawns, resign, refresh the page — everything should survive

## Known limitations

- No HTTPS in dev — JWT cookies are not marked `Secure`; do not deploy as-is
- No rate-limiting on auth endpoints
- Matchmaking queue and pending `match_found` messages don't survive a server restart
- The HTTP `POST /games/:id/move` endpoint doesn't broadcast to the opponent — only the WebSocket path does

## License

MIT
