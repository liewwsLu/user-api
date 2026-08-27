package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/liewwsLu/user-api/internal/config"
	"github.com/liewwsLu/user-api/internal/handlers"
	"github.com/liewwsLu/user-api/internal/storage"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("config error:", err)
		return
	}
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		fmt.Println("open error:", err)
		return
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		fmt.Println("ping db error:", err)
		return
	}
	fmt.Println("successfully connected to database")
	store := storage.NewPostgresStorage(db)
	handler := handlers.New(store)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.HealthHandler)
	mux.HandleFunc("/users", handler.UsersHandler)
	mux.HandleFunc("/user", handler.UserHandler)
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()
	serverErr := make(chan error, 1)
	fmt.Println("starting HTTP server on", server.Addr)
	go func() {
		serverErr <- server.ListenAndServe()
	}()
	select {
	case <-signalCtx.Done():
		fmt.Println("shutdown signal received")
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Println("server error:", err)
		}
		return
	}
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Println("shutdown error:", err)
		return
	}
	fmt.Println("HTTP server stopped")
}
