package storage

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"
	"user-api/internal/models"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func clearUsers(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec("TRUNCATE TABLE users RESTART IDENTITY"); err != nil {
		t.Fatalf("truncate users error: %v", err)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	databaseURL, databaseExists := os.LookupEnv("TEST_DATABASE_URL")
	if !databaseExists {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	databaseURL = strings.TrimSpace(databaseURL)
	if databaseURL == "" {
		t.Fatal("TEST_DATABASE_URL is empty")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("sql.Open() error: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})
	if err := db.Ping(); err != nil {
		t.Fatalf("db.Ping() error: %v", err)
	}
	return db
}

func TestPostgresStorage_CreateUser(t *testing.T) {
	db := openTestDB(t)
	clearUsers(t, db)
	store := NewPostgresStorage(db)
	got, err := store.CreateUser("Egor", "egor@example.com")
	if err != nil {
		t.Fatalf("CreateUser() error: %v", err)
	}
	want := models.User{
		ID:    1,
		Name:  "Egor",
		Email: "egor@example.com",
	}
	if got != want {
		t.Errorf("CreateUser() got: %+v, want: %+v", got, want)
	}

}

func TestPostgresStorage_CreateUser_DuplicateEmail(t *testing.T) {
	db := openTestDB(t)
	clearUsers(t, db)
	store := NewPostgresStorage(db)
	_, err := store.CreateUser("Egor", "egor@example.com")
	if err != nil {
		t.Fatalf("CreateUser() error: %v", err)
	}
	_, err = store.CreateUser("Alex", "egor@example.com")
	if !errors.Is(err, ErrConflict) {
		t.Errorf("CreateUser() error: %v, want: %v", err, ErrConflict)
	}
}

func TestPostgresStorage_FindUserByID(t *testing.T) {
	db := openTestDB(t)
	clearUsers(t, db)
	store := NewPostgresStorage(db)
	want := models.User{
		ID:    1,
		Name:  "Egor",
		Email: "egor@example.com",
	}
	_, err := db.Exec("INSERT INTO users(name,email) VALUES($1,$2)", want.Name, want.Email)
	if err != nil {
		t.Fatalf("db.Exec() error: %v", err)
	}
	got, err := store.FindUserByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("FindUserByID() error: %v", err)
	}
	if got != want {
		t.Errorf("FindUserByID() got: %+v, want: %+v", got, want)
	}
}

func TestPostgresStorage_FindUserByID_NotFound(t *testing.T) {
	db := openTestDB(t)
	clearUsers(t, db)
	store := NewPostgresStorage(db)
	_, err := store.FindUserByID(context.Background(), 999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("FindUserByID() got: %v, want: %v", err, ErrNotFound)
	}
}

func TestPostgresStorage_ListUsers(t *testing.T) {
	db := openTestDB(t)
	clearUsers(t, db)
	store := NewPostgresStorage(db)
	want := []models.User{
		{ID: 1, Name: "Egor", Email: "egor@example.com"},
		{ID: 2, Name: "Alex", Email: "alex@example.com"},
	}
	_, err := db.Exec(
		"INSERT INTO users(name, email) VALUES($1, $2), ($3, $4)",
		want[0].Name,
		want[0].Email,
		want[1].Name,
		want[1].Email,
	)
	if err != nil {
		t.Fatalf("db.Exec() error: %v", err)
	}
	got, err := store.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("ListUsers() error: %v", err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("ListUsers() got: %+v, want: %+v", got, want)
	}
}

func TestPostgresStorage_ListUsers_Empty(t *testing.T) {
	db := openTestDB(t)
	clearUsers(t, db)
	store := NewPostgresStorage(db)
	got, err := store.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("ListUsers() error: %v", err)
	}
	if got == nil {
		t.Error("ListUsers() returned nil slice, want non-nil empty slice")
	}
	if len(got) != 0 {
		t.Errorf("ListUsers() got: %d, want: 0", len(got))
	}
}

func TestPostgresStorage_UpdateUser(t *testing.T) {
	db := openTestDB(t)
	clearUsers(t, db)
	store := NewPostgresStorage(db)
	oldUser := models.User{
		ID:    1,
		Name:  "Egor",
		Email: "old@example.com",
	}
	_, err := db.Exec("INSERT INTO users(name,email) VALUES($1,$2)", oldUser.Name, oldUser.Email)
	if err != nil {
		t.Fatalf("db.Exec() error: %v", err)
	}
	want := models.User{
		ID:    1,
		Name:  "Egor_Updated",
		Email: "new@example.com",
	}
	got, err := store.UpdateUser(1, want.Name, want.Email)
	if err != nil {
		t.Fatalf("UpdateUser() error: %v", err)
	}
	if got != want {
		t.Errorf("UpdateUser() got: %+v,want: %+v", got, want)
	}
}

func TestPostgresStorage_UpdateUser_EmailConflict(t *testing.T) {
	db := openTestDB(t)
	clearUsers(t, db)
	store := NewPostgresStorage(db)
	oldUserF := models.User{
		ID:    1,
		Name:  "Egor",
		Email: "egor@example.com",
	}
	oldUserS := models.User{
		ID:    2,
		Name:  "Alex",
		Email: "alex@example.com",
	}
	_, err := db.Exec("INSERT INTO users(name,email) VALUES($1,$2),($3,$4)", oldUserF.Name, oldUserF.Email, oldUserS.Name, oldUserS.Email)
	if err != nil {
		t.Fatalf("db.Exec() error: %v", err)
	}
	_, err = store.UpdateUser(1, "Egor_Updated", "alex@example.com")
	if !errors.Is(err, ErrConflict) {
		t.Errorf("UpdateUser() error: %v, want: %v", err, ErrConflict)
	}
}

func TestPostgresStorage_UpdateUser_NotFound(t *testing.T) {
	db := openTestDB(t)
	clearUsers(t, db)
	store := NewPostgresStorage(db)
	_, err := store.UpdateUser(999, "Dima", "dima@example.com")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("UpdateUser() got: %v, want: %v", err, ErrNotFound)
	}
}

func TestPostgresStorage_DeleteUser(t *testing.T) {
	db := openTestDB(t)
	clearUsers(t, db)
	store := NewPostgresStorage(db)
	_, err := db.Exec("INSERT INTO users(name,email) VALUES($1,$2)", "Egor", "egor@example.com")
	if err != nil {
		t.Fatalf("db.Exec() error: %v", err)
	}
	if err = store.DeleteUser(1); err != nil {
		t.Fatalf("DeleteUser() error: %v", err)
	}
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE id = $1", 1).Scan(&count)
	if err != nil {
		t.Fatalf("QueryRow() error: %v", err)
	}
	if count != 0 {
		t.Error("DeleteUser wasn't be succesful")
	}
}

func TestPostgresStorage_DeleteUser_NotFound(t *testing.T) {
	db := openTestDB(t)
	clearUsers(t, db)
	store := NewPostgresStorage(db)
	err := store.DeleteUser(999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("DeleteUser() error: %v, want: %v", err, ErrNotFound)
	}
}
