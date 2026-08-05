package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"user-api/internal/models"
)

type fakeUserStorage struct {
	users        []models.User
	listErr      error
	createErr    error
	createdUser  models.User
	gotName      string
	gotEmail     string
	createCalled bool
}

func (f *fakeUserStorage) ListUsers() ([]models.User, error) {
	return f.users, f.listErr
}

func (f *fakeUserStorage) FindUserByID(id int) (models.User, error) {
	panic("not implemented")
}

func (f *fakeUserStorage) CreateUser(name, email string) (models.User, error) {
	f.createCalled = true
	f.gotEmail = email
	f.gotName = name
	return f.createdUser, f.createErr
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

func TestUsersHandlerCreateUser(t *testing.T) {
	wantUser := models.User{
		ID:    1,
		Name:  "Dima",
		Email: "dima@yande.ri",
	}
	fake := &fakeUserStorage{createdUser: wantUser}
	h := New(fake)
	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Dima","email":"dima@yande.ri"}`))
	recorder := httptest.NewRecorder()
	h.UsersHandler(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Errorf("CreateUser() status: %d, want: %d", recorder.Code, http.StatusCreated)
	}
	if !fake.createCalled {
		t.Fatalf("CreateUser was not called")
	}
	if fake.gotName != wantUser.Name {
		t.Errorf(
			"CreateUser() name = %q, want %q",
			fake.gotName,
			wantUser.Name,
		)
	}

	if fake.gotEmail != wantUser.Email {
		t.Errorf(
			"CreateUser() email = %q, want %q",
			fake.gotEmail,
			wantUser.Email,
		)
	}
	var gotUser models.User
	err := json.NewDecoder(recorder.Body).Decode(&gotUser)
	if err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if gotUser.ID != wantUser.ID {
		t.Errorf("response ID = %d, want %d", gotUser.ID, wantUser.ID)
	}
	if gotUser.Name != wantUser.Name {
		t.Errorf("Name() got: %q, want: %q", gotUser.Name, wantUser.Name)
	}
	if gotUser.Email != wantUser.Email {
		t.Errorf("Email() got: %q, want: %q", gotUser.Email, wantUser.Email)
	}
}
