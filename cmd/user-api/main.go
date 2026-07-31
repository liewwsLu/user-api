package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"user-api/internal/handlers"
	"user-api/internal/storage"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	conSQL := "postgres://user:password@localhost:5432/user_api?sslmode=disable"
	bd, err := sql.Open("pgx", conSQL)
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

	http.HandleFunc("/health", h.HealthHandler)
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		h.UsersHandler(w, r)
	})
	http.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		h.UserHandler(w, r)
	})
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
