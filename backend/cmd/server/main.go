package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"krishna.com/go-chess-backend/internal/auth"
	"krishna.com/go-chess-backend/internal/db"
	"krishna.com/go-chess-backend/internal/game"
	"krishna.com/go-chess-backend/internal/handlers"
	"krishna.com/go-chess-backend/internal/matchmaking"
	"krishna.com/go-chess-backend/internal/store"
	"krishna.com/go-chess-backend/internal/ws"
)

var frontendOrigin = "http://localhost:5173"

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", frontendOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// spaHandler serves the built frontend: a real file if it exists, otherwise
// index.html so client-side routes (/games, /play/:id, ...) work on refresh.
func spaHandler(dir string) http.Handler {
	fs := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	if origin := os.Getenv("ALLOWED_ORIGIN"); origin != "" {
		frontendOrigin = origin
	}
	ws.SetAllowedOrigin(frontendOrigin)

	var st store.Store
	var userStore auth.UserStorer
	var dbPing func(context.Context) error

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		pool, err := db.Connect(dbURL)
		if err != nil {
			logger.Fatalf("db connect: %v", err)
		}
		defer pool.Close()

		if err := db.Migrate(pool); err != nil {
			logger.Fatalf("db migrate: %v", err)
		}

		logger.Println("connected to postgres")
		dbPing = pool.Ping
		st = store.NewPostgresStore(pool)
		userStore = auth.NewPostgresUserStore(pool)
	} else {
		logger.Println("DATABASE_URL not set — using in-memory stores")
		st = store.NewMemoryStore()
		userStore = auth.NewUserStore()
	}

	q := matchmaking.NewQueue()
	hub := ws.NewHub()
	manager := game.NewGameManager()
	authHandler := auth.NewHandler(userStore)
	api := handlers.NewApi(st, logger, q, hub, manager)

	// Throttle auth endpoints per client IP to slow down brute-force attempts.
	authLimiter := auth.NewRateLimiter(0.5, 5)

	mux := http.NewServeMux()

	// Auth routes (public)
	mux.Handle("/auth/register", authLimiter.Limit(http.HandlerFunc(authHandler.Register)))
	mux.Handle("/auth/login", authLimiter.Limit(http.HandlerFunc(authHandler.Login)))
	mux.HandleFunc("/auth/logout", authHandler.Logout)
	mux.HandleFunc("/auth/me", authHandler.Me)

	// Game routes
	mux.HandleFunc("/games", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			auth.Middleware(http.HandlerFunc(api.ListGames)).ServeHTTP(w, r)
		case http.MethodPost:
			api.CreateGame(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/games/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if strings.HasSuffix(r.URL.Path, "/moves") {
				auth.Middleware(http.HandlerFunc(api.GetMoves)).ServeHTTP(w, r)
				return
			}
			api.GetGame(w, r)
			return
		}
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/move") {
			auth.Middleware(http.HandlerFunc(api.MakeMove)).ServeHTTP(w, r)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	})

	// Matchmaking (protected)
	mux.Handle("/matchmaking/join", auth.Middleware(http.HandlerFunc(api.JoinMatchMaking)))

	// WebSocket (auth checked inside handler via cookie)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWS(hub, w, r, st, manager)
	})

	// Root mux: health check at the root, the API under /api.
	root := http.NewServeMux()
	root.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if dbPing != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := dbPing(ctx); err != nil {
				http.Error(w, "db unavailable", http.StatusServiceUnavailable)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	root.Handle("/api/", http.StripPrefix("/api", mux))

	// Serve the built frontend (SPA) for everything not under /api or /healthz.
	root.Handle("/", spaHandler(getenv("STATIC_DIR", "frontend/dist")))

	port := getenv("PORT", "8080")

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      corsMiddleware(root),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Printf("listening on %s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	logger.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("shutdown: %v", err)
	}
	logger.Println("server stopped")
}
