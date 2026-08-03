package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"user-api/internal/models"
)

type fakeUserStorage struct {
	users   []models.User
	listErr error
}

func (f *fakeUserStorage) ListUsers() ([]models.User, error) {
	return f.users, f.listErr
}

func (f *fakeUserStorage) FindUserByID(id int) (models.User, error) {
	panic("not implemented")
}

func (f *fakeUserStorage) CreateUser(name, email string) (models.User, error) {
	panic("not implemented")
}

func (f *fakeUserStorage) UpdateUser(id int, name, email string) (models.User, error) {
	panic("not implemented")
}

func (f *fakeUserStorage) DeleteUser(id int) error {
	panic("not implemented")
}

func TestUsersHandlerReturnsUsers(t *testing.T) {
	wantUsers := []models.User{
		{
			ID:    1,
			Name:  "Egor",
			Email: "egor@ewew",
		},
	}
	fake := &fakeUserStorage{
		users: wantUsers,
	}
	h := New(fake)
	request := httptest.NewRequest(http.MethodGet, "/users", nil)
	recorder := httptest.NewRecorder()
	h.UsersHandler(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Errorf(
			"UsersHandler() status = %d, want %d",
			recorder.Code,
			http.StatusOK,
		)
	}
	var gotUsers []models.User
	err := json.NewDecoder(recorder.Body).Decode(&gotUsers)
	if err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if len(gotUsers) != len(wantUsers) {
		t.Fatalf("UsersHandler() users length = %d, want %d", len(gotUsers), len(wantUsers))
	}
	if gotUsers[0] != wantUsers[0] {
		t.Errorf("UsersHandler() first user = %+v, want %+v", gotUsers[0], wantUsers[0])
	}
}

func TestUsersHandlerReturnsInternalServerError(t *testing.T) {
	storageErr := errors.New("database unavailable")
	fake := &fakeUserStorage{
		listErr: storageErr,
	}
	h := New(fake)
	request := httptest.NewRequest(http.MethodGet, "/users", nil)
	recorder := httptest.NewRecorder()
	h.UsersHandler(recorder, request)
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("UsersHandler() status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	var gotErr map[string]string
	err := json.NewDecoder(recorder.Body).Decode(&gotErr)
	if err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if gotErr["error"] != storageErr.Error() {
		t.Errorf("UsersHandler() error = %q, want %q", gotErr["error"], storageErr.Error())
	}
}
