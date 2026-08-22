package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
	"user-api/internal/config"
	"user-api/internal/handlers"
	"user-api/internal/storage"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("config error:", err)
		return
	}
	bd, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		fmt.Println("open error:", err)
		return
	}
	defer bd.Close()
	err = bd.Ping()
	if err != nil {
		fmt.Println("ping db error:", err)
		return
	}
	fmt.Println("succesful connected")
	p := storage.NewPostgresStorage(bd)
	h := handlers.New(p)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.HealthHandler)
	mux.HandleFunc("/users", h.UsersHandler)
	mux.HandleFunc("/user", h.UserHandler)
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	err = server.ListenAndServe()
	if err != nil {
		fmt.Println("Error:", err)
	}
}
