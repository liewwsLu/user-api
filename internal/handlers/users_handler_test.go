package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"user-api/internal/models"
	"user-api/internal/storage"
)

type fakeUserStorage struct {
	users        []models.User
	listErr      error
	createErr    error
	updateErr    error
	findErr      error
	deleteErr    error
	createdUser  models.User
	updatedUser  models.User
	foundUser    models.User
	gotName      string
	gotEmail     string
	gotID        int
	createCalled bool
	updateCalled bool
	findCalled   bool
	deleteCalled bool
}

func (f *fakeUserStorage) ListUsers() ([]models.User, error) {
	return f.users, f.listErr
}

func (f *fakeUserStorage) FindUserByID(id int) (models.User, error) {
	f.findCalled = true
	f.gotID = id
	return f.foundUser, f.findErr
}

func (f *fakeUserStorage) CreateUser(name, email string) (models.User, error) {
	f.createCalled = true
	f.gotEmail = email
	f.gotName = name
	return f.createdUser, f.createErr
}

func (f *fakeUserStorage) UpdateUser(id int, name, email string) (models.User, error) {
	f.updateCalled = true
	f.gotID = id
	f.gotName = name
	f.gotEmail = email
	return f.updatedUser, f.updateErr
}

func (f *fakeUserStorage) DeleteUser(id int) error {
	f.deleteCalled = true
	f.gotID = id
	return f.deleteErr
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
		Email: "dima@yandex.ru",
	}
	fake := &fakeUserStorage{createdUser: wantUser}
	h := New(fake)
	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"     Dima     ","email":"     dima@yandex.ru      "}`))
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

func TestUsersHandlerRejectsInvalidCreateJSON(t *testing.T) {
	tests := []struct {
		name        string
		requestJSON string
	}{
		{name: "malformed JSON", requestJSON: `{"name":`},
		{name: "unknown field", requestJSON: `{"name":"Egor","email":"egor@example.com","role":"admin"}`},
		{
			name: "multiple JSON objects",
			requestJSON: `{"name":"Egor","email":"egor@example.com"}
						{"name":"Dima","email":"dima@example.com"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fake := &fakeUserStorage{}
			h := New(fake)

			request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(test.requestJSON))
			recorder := httptest.NewRecorder()
			h.UsersHandler(recorder, request)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("UsersHandler() status: %d, want: %d", recorder.Code, http.StatusBadRequest)
			}
			if fake.createCalled {
				t.Fatalf("CreateUser() was called")
			}
			var gotErr map[string]string
			err := json.NewDecoder(recorder.Body).Decode(&gotErr)
			if err != nil {
				t.Fatalf("decode response body: %v", err)
			}
			if gotErr["error"] != "invalid json" {
				t.Errorf("json status: %q, want: %q", gotErr["error"], "invalid json")
			}
		})
	}
}

func TestUsersHandlerMapsCreateUserErrors(t *testing.T) {
	tests := []struct {
		name       string
		wantStatus int
		storageErr error
	}{
		{
			name:       "validation error",
			wantStatus: http.StatusBadRequest,
			storageErr: storage.ErrValidation,
		},
		{
			name:       "conflict error",
			wantStatus: http.StatusConflict,
			storageErr: storage.ErrConflict,
		},
		{
			name:       "unexpected error",
			wantStatus: http.StatusInternalServerError,
			storageErr: errors.New("database unavailable"),
		},
	}
	for _, e := range tests {
		t.Run(e.name, func(t *testing.T) {
			fake := &fakeUserStorage{
				createErr: e.storageErr,
			}
			h := New(fake)
			request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Dima","email":"dima@example.com"}`))
			recorder := httptest.NewRecorder()
			h.UsersHandler(recorder, request)
			if !fake.createCalled {
				t.Fatalf("UsersHandler() got: %v, want: %v", fake.createCalled, true)
			}
			if recorder.Code != e.wantStatus {
				t.Errorf("Status() got: %d, want: %d", recorder.Code, e.wantStatus)
			}
			var gotErr map[string]string
			err := json.NewDecoder(recorder.Body).Decode(&gotErr)
			if err != nil {
				t.Fatalf("decode response err: %v", err)
			}
			if gotErr["error"] != e.storageErr.Error() {
				t.Errorf("Error() got: %q, want: %q", gotErr["error"], e.storageErr.Error())
			}
		})
	}
}

func TestUsersHandlerRejectsInvalidCreateUserInput(t *testing.T) {
	tests := []struct {
		name        string
		requestJSON string
		wantErr     string
	}{
		{name: "empty name",
			requestJSON: `{"name":"", "email":" dima@example.com "}`,
			wantErr:     "name is required"},
		{name: "name consists only spaces",
			requestJSON: `{"name":"   ", "email":"dima@example.com"    }`,
			wantErr:     "name is required"},
		{name: "empty email",
			requestJSON: `{"name":"     Dima", "email":""}`,
			wantErr:     "email is required"},
		{name: "email consists only spaces",
			requestJSON: `{"name":"     Dima", "email":"    "}`,
			wantErr:     "email is required"},
		{name: "email without at sign",
			requestJSON: `{"name":"Dima", "email":"hello"}`,
			wantErr:     "email is invalid"},
		{name: "email with display name",
			requestJSON: `{"name":"Dima", "email":"Dima <dima@example.com>"}`,
			wantErr:     "email is invalid"},
		{name: "email without domain",
			requestJSON: `{"name":"Dima", "email":"dima@"}`,
			wantErr:     "email is invalid"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fake := &fakeUserStorage{}
			h := New(fake)
			request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(test.requestJSON))
			recorder := httptest.NewRecorder()
			h.UsersHandler(recorder, request)
			if recorder.Code != http.StatusBadRequest {
				t.Errorf("Status() got: %d, want: %d", recorder.Code, http.StatusBadRequest)
			}
			if fake.createCalled {
				t.Fatalf("CreateUser was called")
			}
			var gotErr map[string]string
			err := json.NewDecoder(recorder.Body).Decode(&gotErr)
			if err != nil {
				t.Fatalf("Decode response error: %v", err)
			}
			if gotErr["error"] != test.wantErr {
				t.Errorf("Error() got: %q, want: %q", gotErr["error"], test.wantErr)
			}

		})
	}
}

func TestValidateUserRequestLength(t *testing.T) {
	tests := []struct {
		name    string
		wantErr string
		request UserRequest
	}{
		{name: "accepts name with 100 runes",
			request: UserRequest{
				Name:  strings.Repeat("Я", 100),
				Email: "dima@example.com",
			},
			wantErr: ""},
		{name: "rejects name with 101 runes",
			request: UserRequest{
				Name:  strings.Repeat("Я", 101),
				Email: "dima@example.com",
			},
			wantErr: "name is too long"},
		{name: "accepts email with 254 bytes",
			request: UserRequest{
				Name: "Dima",
				Email: strings.Repeat("a", 64) + "@" +
					strings.Repeat("b", 63) + "." +
					strings.Repeat("c", 63) + "." +
					strings.Repeat("d", 61),
			},
			wantErr: "",
		},
		{name: "rejects email with 255 bytes",
			request: UserRequest{
				Name: "Dima",
				Email: strings.Repeat("a", 64) + "@" +
					strings.Repeat("b", 63) + "." +
					strings.Repeat("c", 63) + "." +
					strings.Repeat("d", 62),
			},
			wantErr: "email is too long"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := test.request
			err := validateUserRequest(&req)
			if test.wantErr == "" {
				if err != nil {
					t.Fatalf("validateUserRequest() error = %v, want = nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf(
					"validateUserRequest() error = nil, want %q",
					test.wantErr,
				)
			}

			if err.Error() != test.wantErr {
				t.Errorf(
					"validateUserRequest() error = %q, want %q",
					err.Error(),
					test.wantErr,
				)
			}
		})
	}
}

func TestUserHandlerUpdatesUser(t *testing.T) {
	wantUser := models.User{
		ID:    1,
		Name:  "Egor",
		Email: "egor@example.com",
	}
	fake := &fakeUserStorage{
		updatedUser: wantUser,
	}
	h := New(fake)
	request := httptest.NewRequest(http.MethodPut, "/user?id=1", strings.NewReader(`{"name":"   Egor    ","email":"   egor@example.com   "}`))
	recorder := httptest.NewRecorder()
	h.UserHandler(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Errorf("Status() got: %d, want: %d", recorder.Code, http.StatusOK)
	}
	if !fake.updateCalled {
		t.Fatalf("UpdateUser was not called")
	}
	if fake.gotID != 1 {
		t.Errorf("ID() got: %d, want: %d", fake.gotID, 1)
	}
	if fake.gotName != wantUser.Name {
		t.Errorf(
			"UpdateUser() name = %q, want %q",
			fake.gotName,
			wantUser.Name,
		)
	}
	if fake.gotEmail != wantUser.Email {
		t.Errorf(
			"UpdateUser() email = %q, want %q",
			fake.gotEmail,
			wantUser.Email,
		)
	}
	var gotUser models.User
	err := json.NewDecoder(recorder.Body).Decode(&gotUser)
	if err != nil {
		t.Fatalf("Decode response error: %v", err)
	}
	if gotUser != wantUser {
		t.Errorf("User() got: %v, want: %v", gotUser, wantUser)
	}
}

func TestUserHandlerRejectsInvalidUpdateInput(t *testing.T) {
	tests := []struct {
		name        string
		requestJSON string
		wantErr     string
	}{
		{name: "malformed JSON", requestJSON: `{"name":}`, wantErr: "invalid json"},
		{name: "empty name", requestJSON: `{"name":"","email":"egor@example.com"}`, wantErr: "name is required"},
		{name: "invalid email", requestJSON: `{"name":"Egor","email":"hello"}`, wantErr: "email is invalid"},
		{name: "unknown field", requestJSON: `{"name":"Egor","email":"egor@example.com","role":"admin"}`, wantErr: "invalid json"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fake := &fakeUserStorage{}
			h := New(fake)
			request := httptest.NewRequest(http.MethodPut, "/user?id=1", strings.NewReader(test.requestJSON))
			recorder := httptest.NewRecorder()
			h.UserHandler(recorder, request)
			if recorder.Code != http.StatusBadRequest {
				t.Errorf("Status() got: %d, want: %d", recorder.Code, http.StatusBadRequest)
			}
			if fake.updateCalled {
				t.Fatalf("UpdateUser was called")
			}
			var gotErr map[string]string
			err := json.NewDecoder(recorder.Body).Decode(&gotErr)
			if err != nil {
				t.Fatalf("Decode response error: %v", err)
			}
			if gotErr["error"] != test.wantErr {
				t.Errorf("Error() got: %q, want: %q", gotErr["error"], test.wantErr)
			}
		})
	}
}

func TestUserHandlerMapsUpdateUserErrors(t *testing.T) {
	tests := []struct {
		name       string
		wantStatus int
		storageErr error
	}{
		{
			name:       "validation error",
			wantStatus: http.StatusBadRequest,
			storageErr: storage.ErrValidation,
		},
		{
			name:       "not found error",
			wantStatus: http.StatusNotFound,
			storageErr: storage.ErrNotFound,
		},
		{
			name:       "conflict error",
			wantStatus: http.StatusConflict,
			storageErr: storage.ErrConflict,
		},
		{
			name:       "unexpected error",
			wantStatus: http.StatusInternalServerError,
			storageErr: errors.New("database unavailable"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fake := &fakeUserStorage{
				updateErr: test.storageErr,
			}
			h := New(fake)
			request := httptest.NewRequest(http.MethodPut, "/user?id=1", strings.NewReader(`{"name":"Dima","email":"dima@example.com"}`))
			recorder := httptest.NewRecorder()
			h.UserHandler(recorder, request)
			if !fake.updateCalled {
				t.Fatalf("UserHandler() got: %v, want: %v", fake.updateCalled, true)
			}
			if recorder.Code != test.wantStatus {
				t.Errorf("Status() got: %d, want: %d", recorder.Code, test.wantStatus)
			}
			var gotErr map[string]string
			err := json.NewDecoder(recorder.Body).Decode(&gotErr)
			if err != nil {
				t.Fatalf("decode response err: %v", err)
			}
			if gotErr["error"] != test.storageErr.Error() {
				t.Errorf("Error() got: %q, want: %q", gotErr["error"], test.storageErr.Error())
			}
		})
	}
}

func TestUserHandlerReturnsUserByID(t *testing.T) {
	wantUser := models.User{
		ID:    7,
		Name:  "Egor",
		Email: "egor@example.com",
	}

	fake := &fakeUserStorage{
		foundUser: wantUser,
	}
	h := New(fake)

	request := httptest.NewRequest(http.MethodGet, "/user?id=7", nil)
	recorder := httptest.NewRecorder()

	h.UserHandler(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"UserHandler() status = %d, want %d",
			recorder.Code,
			http.StatusOK,
		)
	}

	if !fake.findCalled {
		t.Fatalf("FindUserByID() was not called")
	}

	if fake.gotID != wantUser.ID {
		t.Errorf(
			"FindUserByID() ID = %d, want %d",
			fake.gotID,
			wantUser.ID,
		)
	}

	var gotUser models.User
	err := json.NewDecoder(recorder.Body).Decode(&gotUser)
	if err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if gotUser != wantUser {
		t.Errorf(
			"UserHandler() user = %+v, want %+v",
			gotUser,
			wantUser,
		)
	}
}

func TestUserHandlerRejectsInvalidID(t *testing.T) {
	fake := &fakeUserStorage{}
	h := New(fake)

	request := httptest.NewRequest(http.MethodGet, "/user?id=0", nil)
	recorder := httptest.NewRecorder()

	h.UserHandler(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"UserHandler() status = %d, want %d",
			recorder.Code,
			http.StatusBadRequest,
		)
	}

	if fake.findCalled {
		t.Fatalf("FindUserByID() was called")
	}

	var gotErr map[string]string
	err := json.NewDecoder(recorder.Body).Decode(&gotErr)
	if err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if gotErr["error"] != "invalid id" {
		t.Errorf(
			"UserHandler() error = %q, want %q",
			gotErr["error"],
			"invalid id",
		)
	}
}

func TestUserHandlerMapsFindUserErrors(t *testing.T) {
	tests := []struct {
		name       string
		storageErr error
		wantStatus int
	}{
		{
			name:       "user not found",
			storageErr: storage.ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "unexpected storage error",
			storageErr: errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fake := &fakeUserStorage{
				findErr: test.storageErr,
			}
			h := New(fake)

			request := httptest.NewRequest(http.MethodGet, "/user?id=7", nil)
			recorder := httptest.NewRecorder()

			h.UserHandler(recorder, request)

			if !fake.findCalled {
				t.Fatalf("FindUserByID() was not called")
			}

			if fake.gotID != 7 {
				t.Errorf(
					"FindUserByID() ID = %d, want %d",
					fake.gotID,
					7,
				)
			}

			if recorder.Code != test.wantStatus {
				t.Fatalf(
					"UserHandler() status = %d, want %d",
					recorder.Code,
					test.wantStatus,
				)
			}

			var gotErr map[string]string
			err := json.NewDecoder(recorder.Body).Decode(&gotErr)
			if err != nil {
				t.Fatalf("decode response body: %v", err)
			}

			if gotErr["error"] != test.storageErr.Error() {
				t.Errorf(
					"UserHandler() error = %q, want %q",
					gotErr["error"],
					test.storageErr.Error(),
				)
			}
		})
	}
}

func TestUserHandlerDeletesUser(t *testing.T) {
	fake := &fakeUserStorage{}
	h := New(fake)
	request := httptest.NewRequest(http.MethodDelete, "/user?id=1", nil)
	recorder := httptest.NewRecorder()
	h.UserHandler(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("Status() got: %d, want: %d", recorder.Code, http.StatusOK)
	}
	if !fake.deleteCalled {
		t.Fatalf("DeleteUser() was not called")
	}
	if fake.gotID != 1 {
		t.Errorf("DeleteUser() id: %d, want: %d", fake.gotID, 1)
	}
	var gotStatus map[string]string
	err := json.NewDecoder(recorder.Body).Decode(&gotStatus)
	if err != nil {
		t.Fatalf("decode response error: %v", err)
	}
	if gotStatus["status"] != "deleted" {
		t.Errorf("DeleteUser() status: %q, want: %q", gotStatus["status"], "deleted")
	}
}

func TestUserHandlerRejectsInvalidDeleteID(t *testing.T) {
	fake := &fakeUserStorage{}
	h := New(fake)
	request := httptest.NewRequest(http.MethodDelete, "/user?id=0", nil)
	recorder := httptest.NewRecorder()
	h.UserHandler(recorder, request)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("Status() got: %d, want: %d", recorder.Code, http.StatusBadRequest)
	}
	if fake.deleteCalled {
		t.Fatalf("DeleteUser() was called")
	}
	var gotErr map[string]string
	err := json.NewDecoder(recorder.Body).Decode(&gotErr)
	if err != nil {
		t.Fatalf("decode response error: %v", err)
	}
	if gotErr["error"] != "invalid id" {
		t.Errorf("DeleteUser() error: %q, want: %q", gotErr["error"], "invalid id")
	}
}

func TestUserHandlerMapsDeleteUserErrors(t *testing.T) {
	tests := []struct {
		name       string
		errStorage error
		wantStatus int
	}{
		{name: "return error when user not found",
			errStorage: storage.ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "return error when unexpected error",
			errStorage: errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fake := &fakeUserStorage{
				deleteErr: test.errStorage,
			}
			h := New(fake)
			request := httptest.NewRequest(http.MethodDelete, "/user?id=50", nil)
			recorder := httptest.NewRecorder()
			h.UserHandler(recorder, request)
			if recorder.Code != test.wantStatus {
				t.Fatalf("Status() got: %d, want: %d", recorder.Code, test.wantStatus)
			}
			if !fake.deleteCalled {
				t.Fatalf("DeleteUser() was not called")
			}
			if fake.gotID != 50 {
				t.Errorf("DeleteUser() ID = %d, want %d", fake.gotID, 50)
			}
			var gotErr map[string]string
			err := json.NewDecoder(recorder.Body).Decode(&gotErr)
			if err != nil {
				t.Fatalf("decode response error: %v", err)
			}
			if gotErr["error"] != test.errStorage.Error() {
				t.Errorf("DeleteUser() error: %q, want: %q", gotErr["error"], test.errStorage.Error())
			}
		})
	}
}

func TestUserHandlerRejectsUnsupportedMethod(t *testing.T) {
	fake := &fakeUserStorage{}
	h := New(fake)
	request := httptest.NewRequest(http.MethodPost, "/user", nil)
	recorder := httptest.NewRecorder()
	h.UserHandler(recorder, request)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("UserHandler() status: %d, want: %d", recorder.Code, http.StatusMethodNotAllowed)
	}
	var gotErr map[string]string
	err := json.NewDecoder(recorder.Body).Decode(&gotErr)
	if err != nil {
		t.Fatalf("decode response error: %v", err)
	}
	if gotErr["error"] != "method not allowed" {
		t.Errorf("UserHandler() error: %q, want: %q", gotErr["error"], "method not allowed")
	}
}

func TestUsersHandlerRejectsUnsupportedMethod(t *testing.T) {
	fake := &fakeUserStorage{}
	h := New(fake)
	request := httptest.NewRequest(http.MethodPut, "/users", nil)
	recorder := httptest.NewRecorder()
	h.UsersHandler(recorder, request)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("UsersHandler() status: %d, want: %d", recorder.Code, http.StatusMethodNotAllowed)
	}
	var gotErr map[string]string
	err := json.NewDecoder(recorder.Body).Decode(&gotErr)
	if err != nil {
		t.Fatalf("decode response error: %v", err)
	}
	if gotErr["error"] != "method not allowed" {
		t.Errorf("UsersHandler() error: %q, want: %q", gotErr["error"], "method not allowed")
	}
}
