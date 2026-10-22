# chess-arena — Learning Roadmap

A guided path through this codebase, from "what does it do?" to "I could rebuild it."
About 3,000 lines of real code: ~1,900 Go (backend) and ~1,100 TypeScript/React (frontend).

---

## How to use this roadmap

1. **Go in order.** Each stage builds on the one before it. Stage 6 (WebSockets) will make
   little sense without Stage 5 (matchmaking), and so on.
2. **For each stage, do four things:**
   - **Read** the files listed, *in the order listed*. Small files first, glue code last.
   - **Understand** the "Key ideas". These are the concepts the stage is really teaching.
   - **Check yourself** with the questions. If you can answer them without looking, move on.
     If not, re-read, or ask Claude: *"Stage 4, question 2: I don't get it"*.
   - **Try it** with the hands-on exercise. Breaking and fixing things is the fastest way to learn.
3. **Tick the boxes** (`[ ]` → `[x]`) in this file as you finish things so you can see your progress.
4. **Keep the app running** while you read (Stage 0). Watching logs while clicking around
   makes the code real.
5. **"Look closer" notes** point out spots where the code has a quirk, a limitation, or a
   design trade-off. They're good discussion topics and possible improvements, not things you
   need to fix right now.
6. **Budget:** about 1–2 hours per stage. Stages 6 and 7 are the heart of the project, so give
   them the most time.

### Progress tracker

- [ ] Stage 0: Big picture & running it
- [ ] Stage 1: The backend entry point (`main.go`)
- [ ] Stage 2: Authentication (backend + frontend)
- [ ] Stage 3: Data layer: models, stores, Postgres, migrations
- [ ] Stage 4: The chess engine wrapper
- [ ] Stage 5: Matchmaking
- [ ] Stage 6: WebSockets on the server (the Hub)
- [ ] Stage 7: The game page on the client (moves, reconnect, resync)
- [ ] Stage 8: History & replay
- [ ] Stage 9: Build & deployment
- [ ] Stage 10: Capstone: trace everything end-to-end

---

## Stage 0: Big picture & running it

**Goal:** Know what the app does and see it work before reading any code.

**Read:**
- `README.md`: features, endpoint table, "Architecture at a glance"
- `dev.sh`: the one-command dev runner

**Key ideas:**
- Two programs: a **Go server** (`backend/`) and a **React single-page app** (`frontend/`).
- They talk two ways: **HTTP (REST)** for request/response things (login, join queue, fetch a game)
  and **WebSocket** for push things (opponent's move, match found, game over).
- **Postgres** stores users, games, and moves. Without `DATABASE_URL`, the server keeps everything
  in memory instead.

```
 Browser (React)                     Go server                        Postgres
 ───────────────                     ─────────                        ────────
  axios  ── HTTP /api/... ────────▶  http.ServeMux ─▶ handlers ─▶ store ─▶ users/games/moves
  WebSocket ── /api/ws ◀═══════════▶  ws.Hub (lobby + game conns)
```

**Try it (live site):** the app is deployed, so start there.
- [ ] Open your live URL in a normal window and an incognito window.
- [ ] Register two users, click **Play** in both, play a few moves, refresh mid-game, resign.
- [ ] Visit `<your-url>/healthz`. You should see `ok`, which means the server is up and can reach Postgres.

**Try it (local):** you'll need this for every "break it and watch" exercise later, because you
can't safely experiment on the live site.
- [ ] Run `./dev.sh`, open http://localhost:5173 in two browsers, and repeat the steps above.
- [ ] Watch the backend terminal output while you do it.

**Check yourself:**
1. Why do you need two *different* browsers (or incognito) to test with two players?
2. Which parts of the app use HTTP and which use WebSocket?

---

## Stage 1: The backend entry point

**Goal:** Understand how the Go server starts and how a request finds its handler.

**Read:**
- `backend/go.mod`: module name (`krishna.com/go-chess-backend`) and dependencies
- `backend/cmd/server/main.go`: read it top to bottom

**Key ideas:**
- **Dependency wiring in `main`:** every component (store, queue, hub, manager, handlers) is
  created once and passed in. No globals for shared state (except a few config vars).
- **Choosing an implementation at runtime:** `store.Store` is an *interface*; `main` picks
  `PostgresStore` or `MemoryStore` based on `DATABASE_URL`.
- **Two muxes:** the inner `mux` holds API routes like `/games`; the `root` mux mounts it under
  `/api` with `http.StripPrefix`, serves `/healthz`, and serves the built frontend for everything else.
- **Middleware pattern:** `corsMiddleware(root)` and `auth.Middleware(handler)` wrap a handler
  in another handler.
- **SPA fallback (`spaHandler`):** unknown paths return `index.html` so `/play/123` works on refresh.
- **Graceful shutdown:** listen in a goroutine, block on Ctrl+C, then `srv.Shutdown`.

**Check yourself:**
1. A request for `/api/games/abc/moves` arrives. Walk through exactly which code handles it.
2. Why do `GET /games/:id` and `GET /games/:id/moves` share one `HandleFunc("/games/", ...)`?
3. What happens if `JWT_SECRET` is missing? (Hint: look at `auth/jwt.go`. It runs at package init.)

**Try it:**
- [ ] `curl -i localhost:8080/healthz`, then stop Postgres (`docker compose stop`) and curl again.

---

## Stage 2: Authentication

**Goal:** Understand how a user proves who they are on every request.

**Read (backend):**
1. `backend/internal/auth/jwt.go`: token create/parse, cookie set/clear
2. `backend/internal/auth/middleware.go`: puts `userID` into the request context
3. `backend/internal/auth/store.go`: in-memory user store + bcrypt
4. `backend/internal/auth/postgres_store.go`: same interface, Postgres-backed
5. `backend/internal/auth/handler.go`: register / login / logout / me
6. `backend/internal/auth/ratelimit.go`: token-bucket limiter per IP

**Read (frontend):**
1. `frontend/src/context/AuthContext.tsx`: the shared `api` axios client + `useAuth()`
2. `frontend/src/components/ProtectedRoute.tsx`
3. `frontend/src/components/Login.tsx`, `Register.tsx`, `Navbar.tsx`

**Key ideas:**
- **JWT in an HttpOnly cookie:** the browser sends it automatically; JavaScript can't read it
  (protects against XSS token theft). `withCredentials: true` on axios makes cookies flow.
- **Stateless sessions:** the server keeps no session table. It just verifies the signature.
- **Never trust the client for identity:** `playerID` always comes from the verified token,
  never from a request body.
- **bcrypt:** passwords are stored as slow salted hashes.
- **`/auth/me`:** on page load, the React app asks "who am I?" to restore the session.
- **Token bucket:** `NewRateLimiter(0.5, 5)` means a burst of 5, then 1 request every 2 seconds.

**Check yourself:**
1. Why does `ParseToken` check the signing method is HMAC?
2. Why is `Secure` on the cookie controlled by `COOKIE_SECURE` instead of checking `r.TLS`?
3. Where does `UserIDFromContext` get its value, and what happens if a route forgot the middleware?

**Try it:**
- [ ] Log in, open DevTools → Application → Cookies, and paste the `auth_token` into jwt.io. What's inside?
- [ ] Hit `/api/auth/login` with a wrong password 7 times quickly using curl. What changes?

---

## Stage 3: Data layer

**Goal:** Understand what is stored, where, and how the two store implementations differ.

**Read:**
1. `backend/internal/models/game.go`: the `Game` and `MoveRecord` shapes (these become JSON)
2. `backend/internal/db/migrations/001..003_*.sql`: the three tables
3. `backend/internal/db/db.go`: connection pool + a tiny home-made migration runner (`embed.FS`)
4. `backend/internal/store/memory.go`: start with the `Store` interface at the top
5. `backend/internal/store/postgres.go`: focus on `getOrLoadSession`, `ApplyMove`, `GetGame`

**Key ideas:**
- **Hybrid store:** live chess state (`game.Session`) is kept **in memory** for fast move
  validation, while every move is **also written to Postgres**.
- **Lazy rebuild after restart:** `getOrLoadSession` replays the `moves` table into a fresh
  session the first time a game is touched. That's why games survive a server restart.
- **Migrations:** each `.sql` file runs once, and the run is recorded in `schema_migrations`.
- **Interface = swappable backend:** handlers only know `store.Store`.

**Look closer:**
- `ApplyMove` updates `games` and inserts into `moves` as two separate statements, with no
  transaction. What happens if the second one fails?
- `GetGame` returns `ErrorNotFound` for *any* query error, not just "no rows".

**Check yourself:**
1. Draw the three tables and their foreign keys.
2. Server restarts mid-game. The next move arrives. Walk through what `getOrLoadSession` does.
3. Why does `getOrLoadSession` re-check the map after taking the write lock?

**Try it:**
- [ ] `docker compose exec postgres psql -U gochess` and run `SELECT * FROM moves ORDER BY id DESC LIMIT 10;` after playing.

---

## Stage 4: The chess engine wrapper

**Goal:** Understand how chess rules are enforced on the server.

**Read:**
1. `backend/internal/game/session.go`: wraps `notnil/chess`
2. `backend/internal/game/manager.go`: in-memory "who's in which game, with which color"

**Key ideas:**
- The **server is the referee.** `Session.ApplyMove` rejects illegal moves; the client's
  `chess.js` validation is only for good UX.
- Moves travel as **SAN** (e.g. `Nf3`, `exd5`, `e8=Q`), which `MoveStr` parses.
- `FEN` = a snapshot of the board; the move list = the history.
- **Mutex per session** because two players' goroutines can touch the same game at once.
- `GameManager` assigns random colors and maps player → game.

**Look closer:**
- `Turn()` is computed from move count, not from `game.Position().Turn()`. When does that matter?
- `GameManager` is memory-only, so after a restart other code falls back to the store (see `playerColor` in Stage 6).

**Check yourself:**
1. Where exactly does an illegal move get rejected?
2. What's the difference between `Session` and `GameManager.Game`?

---

## Stage 5: Matchmaking

**Goal:** Understand how two strangers get paired and both learn their color.

**Read:**
1. `backend/internal/matchmaking/queue.go`: a FIFO slice behind a mutex
2. `backend/internal/handlers/api.go`: `JoinMatchMaking`
3. `frontend/src/components/MatchMaking.tsx`
4. `frontend/src/context/WSContext.tsx`: the always-on **lobby** WebSocket

**Key ideas: the two-phase color delivery.**
- Player A joins → queue is empty → A waits (HTTP returns `waiting`).
- Player B joins → paired with A → the server:
  - sends `match_found` to **A over the lobby WS** (A's HTTP request already returned), and
  - returns `matched` + color to **B in the HTTP response** (and also over WS).
- Both clients save their color in `localStorage` as `color_<gameId>` and navigate to `/play/<gameId>`.
- If A's lobby socket isn't connected at that moment, the Hub stores the message as **pending**
  and delivers it on connect (Stage 6).

**Look closer:**
- What if a player closes the tab while waiting? Is `LeaveQueue` ever called?
- The queue lives in memory. What does a server restart do to waiting players?

**Check yourself:**
1. Why can't A simply learn about the match from its HTTP response?
2. Where does the client store its color, and what breaks if `localStorage` is cleared mid-game?

---

## Stage 6: WebSockets on the server (the core)

**Goal:** Understand real-time messaging: connection setup, routing, and keep-alive.

**Read:**
1. `backend/internal/ws/hub.go`: `Message`, the `clients` and `pending` maps
2. `backend/internal/ws/handler.go`: `ServeWS` → `serveLobby` / `serveGame`

**Key ideas:**
- **Auth before upgrade:** the cookie is checked *before* the HTTP connection becomes a WebSocket.
  `CheckOrigin` blocks other websites from opening sockets with your cookie.
- **First message is `auth`:** `{type:"auth", payload:{}}` → lobby connection;
  `{type:"auth", payload:{gameId}}` → game connection.
- **Hub keys:** lobby = `playerID`, game = `playerID:gameID`. One player can have both open at once.
  A new connection with the same key closes the old one.
- **Pending messages:** if the recipient isn't connected, the message is parked and sent on register.
- **Move flow in `serveGame`:** parse → find color → check turn → `st.ApplyMove` (legality +
  persist) → forward to opponent → if game ended, broadcast `game_over` to both.
- **Ping/pong:** server pings every 45s and drops the connection if there's no pong within 70s.
  This keeps proxies like Cloudflare from killing idle sockets.
- **Read limits:** 1 KB for the handshake, 4 KB for game messages, so a client can't send huge payloads.

**Look closer:**
- gorilla/websocket allows **only one concurrent writer per connection**. `sendTo` calls
  `conn.WriteJSON` outside the lock, and two goroutines (both players' handlers) can write to the
  same conn, e.g. both sending `game_over`. Is that safe? A common fix is one writer goroutine
  with a send channel per connection.
- `pending` holds only **one** message per key. What if two arrive while someone is offline?
- A rejected move is silently `continue`d. The sender never learns it was rejected.

**Check yourself:**
1. Sequence: White plays `e4`. List every function call on the server, from `ReadJSON` to the
   opponent's socket.
2. Why is `playerID` taken from the cookie and not from `MovePayload.PlayerId`?
3. What happens to a player's game connection when they open the same game in a second tab?

**Try it:**
- [ ] In DevTools → Network → WS, select the `/api/ws` connection and watch frames while you play.

---

## Stage 7: The game page on the client

**Goal:** Understand how the browser shows the board, sends moves, and survives disconnects.

**Read:**
1. `frontend/src/lib/wsUrl.ts`: builds `ws://` or `wss://` from the current page
2. `frontend/src/lib/reconnectingWebSocket.ts`: exponential backoff (500ms → 30s)
3. `frontend/src/hooks/useChessGame.ts`: a `chess.js` instance in a `useRef` + React state for FEN/PGN/status
4. `frontend/src/components/ChessBoard.tsx`: wraps `chessground` (the board UI); `buildDests` = legal moves
5. `frontend/src/components/Play.tsx`: ties it all together
6. `GameInfo.tsx`, `GameControls.tsx`: small presentational pieces

**Key ideas:**
- **Optimistic local move:** your move is applied locally first, then sent to the server.
- **Opponent's move** arrives as `{type:"move", payload:{from,to,promotion}}` and is applied locally.
- **Resync on (re)connect:** `onOpen` → `GET /games/:id` → replay any moves the client is missing.
  This is safe to repeat because it only applies moves past the local count.
- **Promotion:** detect pawn → last rank, show a picker, then send with the chosen piece.
- **Orientation:** comes from the color saved during matchmaking.
- **`useRef` vs `useState`:** the `Chess` object is mutable and lives in a ref; React state holds
  derived values to trigger re-renders.

**Look closer:**
- If the server rejects your move, your local board still shows it (no rollback).
- **Undo** and **Reset** in `GameControls` only change your local board in a live game.
  Should they exist on this page?

**Check yourself:**
1. Wi-Fi drops for 10 seconds while your opponent moves twice. Explain exactly how your board catches up.
2. Why does `Play.tsx` open its own WebSocket instead of reusing the lobby one from `WSContext`?
3. How does `ChessBoard` stop you from moving your opponent's pieces?

**Try it:**
- [ ] Mid-game, stop the backend for ~5 seconds and restart it. Watch the console for reconnect
  logs and check the board recovers.

---

## Stage 8: History & replay

**Goal:** Understand how finished games are listed and stepped through.

**Read:**
1. `backend/internal/store/postgres.go`: `ListGames` (joins usernames)
2. `frontend/src/components/GamesList.tsx`
3. `frontend/src/lib/pgnDisplay.ts`
4. `frontend/src/components/Replay.tsx`: replays moves into a fresh `Chess` up to `currentPly`, with keyboard shortcuts

**Key ideas:**
- Replay needs no special server support: it reuses the stored move list.
- The board is `viewOnly` in replay mode.

**Check yourself:**
1. How does Replay compute the position at ply 17?
2. Could anyone view any game's replay? (Look at which endpoints require auth in `main.go`.)

---

## Stage 9: Build & deployment

**Goal:** Understand how this runs on a laptop vs. in production, and be able to explain
exactly how *your* live deployment works.

**First, figure out which setup is live.** The repo contains two:
- **Docker / cloud** (newest commit, `a63f7d6`): one container, where the Go binary serves the
  API, the WebSocket, and the built frontend. Postgres is a separate service.
- **Self-hosted** (`8f5283e`): `prod.sh` on your own machine. Caddy serves the frontend and
  proxies `/api`, and a Cloudflare tunnel exposes it to the internet.

Write down which one you're on, where Postgres lives, and where the env vars are set.

**Read:**
1. `frontend/vite.config.ts`: dev proxy `/api` → `:8080` (with `ws: true`)
2. `docker-compose.yml`: Postgres only, bound to localhost
3. `.env.example`, `backend/configs/.env.example`, `frontend/.example.env`
4. `Dockerfile`: 3-stage build (Node builds the frontend → Go builds a static binary → tiny Alpine image)
5. `prod.sh`, `Caddyfile`, `deploy/cloudflared.yml.example`

**Key ideas:**
- **Same origin everywhere:** the frontend always calls `/api/...` on its own host. In dev, Vite
  proxies it; in prod, Caddy or the Go server itself does. This is why CORS barely matters and cookies just work.
- **Two prod options:** (a) the `Dockerfile`, where the Go binary serves both API and SPA via
  `spaHandler`; (b) `prod.sh`, where Caddy serves the SPA and proxies `/api`, behind a Cloudflare tunnel.
- **TLS terminates at Cloudflare**, which is why `COOKIE_SECURE` is an env var.
- **Config through env:** `DATABASE_URL`, `JWT_SECRET`, `COOKIE_SECURE`, `ALLOWED_ORIGIN`, `PORT`, `STATIC_DIR`.

**Check yourself:**
1. In dev, the browser is on `:5173` but the API is on `:8080`. Why don't cookies or CORS break?
2. What would go wrong if `ALLOWED_ORIGIN` didn't match the real site URL in production?

3. How does a WebSocket survive the trip through your host's proxy? Why does the server ping
   every 45s? (See `pingLoop` in `ws/handler.go`.)
4. If your host restarts the container mid-game, what is lost and what survives? (Combine Stage 3
   and the "Known limitations" in `README.md`.)

**Try it:**
- [ ] `docker build -t chess-arena .` and run it with `-e JWT_SECRET=... -p 8080:8080`, then
  open http://localhost:8080. This is the same image your host runs.
- [ ] On the live site, open DevTools: check the WebSocket URL is `wss://` (not `ws://`) and the
  `auth_token` cookie has the `Secure` flag. If either is wrong, revisit `COOKIE_SECURE` / `wsUrl.ts`.
- [ ] Find where your host shows server logs. Play a move on the live site and find the log line.

---

## Stage 10: Capstone

**Goal:** Prove you understand the whole system.

- [ ] **Trace a game end-to-end** on paper: register → login → join queue (both players) →
  match found → open game WS → move → checkmate → `game_over` → appears in My Games → replay.
  Name the file and function at every step.
- [ ] **Explain the "Known limitations"** in `README.md`: why each one exists and how you'd fix it.
- [ ] **Pick one "Look closer" note** from Stages 3–7 and implement a fix. Good first ones:
  - send an error message back to the client when a move is rejected
  - one writer goroutine per WebSocket connection (Stage 6)
  - call `LeaveQueue` when a waiting player disconnects
  - wrap `ApplyMove`'s two SQL statements in a transaction
- [ ] **Add a feature:** a chess clock, draw offers, spectators, or a rematch button. Then list
  every file you'd need to touch before you start.
- [ ] **Write tests:** the backend has none yet. `matchmaking/queue.go` and `game/session.go`
  are the easiest places to start (`go test ./...`).

When you can do the first two boxes from memory, you understand this project.
