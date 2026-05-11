package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"krishna.com/go-chess-backend/internal/game"
	"krishna.com/go-chess-backend/internal/handlers"
	"krishna.com/go-chess-backend/internal/matchmaking"
	"krishna.com/go-chess-backend/internal/store"
	"krishna.com/go-chess-backend/internal/ws"
)

type Game struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight request
		if r.Method == http.MethodOptions {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	st := store.NewMemoryStore()
	q := matchmaking.NewQueue()
	hub := ws.NewHub()
	manager := game.NewGameManager()

	api := handlers.NewApi(st, logger, q, hub, manager)
	mux := http.NewServeMux()
	mux.HandleFunc("/games", api.CreateGame)
	mux.HandleFunc("/games/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			api.GetGame(w, r)
			return
		}
		if r.Method == http.MethodPost && len(r.URL.Path) > 1 && strings.HasSuffix(r.URL.Path, "/move") {
			api.MakeMove(w, r)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	})
	mux.HandleFunc("/matchmaking/join", api.JoinMatchMaking)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWS(hub, w, r, st, manager)
	})

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      corsMiddleware(mux),
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
		logger.Fatalf("shutdown:  %v", err)
	}
	logger.Println("server stopped")
}
