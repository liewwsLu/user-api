package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"user-api/internal/models"

	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{
		db: db,
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (s *PostgresStorage) CreateUser(ctx context.Context, name, email string) (models.User, error) {
	if err := ctx.Err(); err != nil {
		return models.User{}, err
	}
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	if name == "" || email == "" {
		return models.User{}, ErrValidation
	}
	u := models.User{}
	err := s.db.QueryRowContext(ctx,
		"INSERT INTO users(name, email) VALUES($1, $2) RETURNING id, name, email",
		name,
		email,
	).Scan(&u.ID, &u.Name, &u.Email)
	if isUniqueViolation(err) {
		return models.User{}, ErrConflict
	}
	if err != nil {
		return models.User{}, fmt.Errorf("create error: %w", err)
	}
	return u, nil

}

func (s *PostgresStorage) FindUserByID(ctx context.Context, id int) (models.User, error) {
	u := models.User{}
	err := s.db.QueryRowContext(ctx,
		"SELECT id, name, email FROM users WHERE id = $1",
		id,
	).Scan(&u.ID, &u.Name, &u.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	if err != nil {
		return models.User{}, fmt.Errorf("find user: %w", err)
	}
	return u, nil
}

func (s *PostgresStorage) ListUsers(ctx context.Context) ([]models.User, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, name, email FROM users ORDER BY 1",
	)
	if err != nil {
		return []models.User{}, fmt.Errorf("db error: %w", err)
	}
	defer rows.Close()
	users := []models.User{}
	for rows.Next() {
		u := models.User{}
		err = rows.Scan(&u.ID, &u.Name, &u.Email)
		if err != nil {
			return []models.User{}, fmt.Errorf("scan error: %w", err)
		}
		users = append(users, u)
	}
	err = rows.Err()
	if err != nil {
		return []models.User{}, fmt.Errorf("read error: %w", err)
	}
	return users, nil
}
func (s *PostgresStorage) DeleteUser(id int) error {
	result, err := s.db.Exec(
		"DELETE FROM users WHERE id = $1",
		id,
	)
	if err != nil {
		return fmt.Errorf("SQL-question error: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("driver error: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStorage) UpdateUser(id int, name, email string) (models.User, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	if name == "" || email == "" {
		return models.User{}, ErrValidation
	}
	u := models.User{}
	err := s.db.QueryRow(
		"UPDATE users SET name = $1, email = $2 WHERE id = $3 RETURNING id, name, email",
		name,
		email,
		id,
	).Scan(&u.ID, &u.Name, &u.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	if isUniqueViolation(err) {
		return models.User{}, ErrConflict
	}
	if err != nil {
		return models.User{}, fmt.Errorf("update error: %w", err)
	}
	return u, nil
}
