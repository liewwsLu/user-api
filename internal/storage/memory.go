package storage

import (
	"context"
	"strings"
	"user-api/internal/models"
)

type MemoryStorage struct {
	users  map[int]models.User
	nextID int
}

func NewMemoryStorage() *MemoryStorage {
	t := MemoryStorage{
		users:  make(map[int]models.User),
		nextID: 1,
	}
	return &t
}

func (s *MemoryStorage) CreateUser(name, email string) (models.User, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	if name == "" || email == "" {
		return models.User{}, ErrValidation
	}
	for _, u := range s.users {
		if u.Email == email {
			return models.User{}, ErrConflict
		}
	}
	new := models.User{
		ID:    s.nextID,
		Name:  name,
		Email: email,
	}
	s.users[s.nextID] = new
	s.nextID++

	return new, nil
}

func (s *MemoryStorage) FindUserByID(ctx context.Context, id int) (models.User, error) {
	if err := ctx.Err(); err != nil {
		return models.User{}, err
	}
	_, ok := s.users[id]
	if ok {
		return s.users[id], nil
	}
	return models.User{}, ErrNotFound
}

func (s *MemoryStorage) ListUsers(ctx context.Context) ([]models.User, error) {
	if err := ctx.Err(); err != nil {
		return []models.User{}, err
	}
	slice := make([]models.User, 0, len(s.users))
	for _, u := range s.users {
		slice = append(slice, u)
	}
	return slice, nil
}

func (s *MemoryStorage) DeleteUser(id int) error {
	_, ok := s.users[id]
	if ok {
		delete(s.users, id)
		return nil
	}
	return ErrNotFound
}

func (s *MemoryStorage) UpdateUser(id int, name, email string) (models.User, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)
	if name == "" || email == "" {
		return models.User{}, ErrValidation
	}
	_, ok := s.users[id]
	if !ok {
		return models.User{}, ErrNotFound
	}
	for _, u := range s.users {
		if u.ID != id && u.Email == email {
			return models.User{}, ErrConflict
		}
	}
	u := models.User{
		ID:    id,
		Name:  name,
		Email: email,
	}
	s.users[id] = u
	return u, nil
}
