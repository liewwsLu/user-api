package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"unicode/utf8"
	"user-api/internal/helper"
	"user-api/internal/models"
	"user-api/internal/storage"
)

type UserStorage interface {
	ListUsers(ctx context.Context) ([]models.User, error)
	FindUserByID(ctx context.Context, id int) (models.User, error)
	CreateUser(ctx context.Context, name, email string) (models.User, error)
	UpdateUser(ctx context.Context, id int, name, email string) (models.User, error)
	DeleteUser(id int) error
}

type Handler struct {
	store UserStorage
}

func New(store UserStorage) *Handler {
	return &Handler{
		store: store,
	}
}

type UserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func validateUserRequest(req *UserRequest) error {
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.Email == "" {
		return errors.New("email is required")
	}
	if utf8.RuneCountInString(req.Name) > 100 {
		return errors.New("name is too long")
	}
	if len(req.Email) > 254 {
		return errors.New("email is too long")
	}
	address, err := mail.ParseAddress(req.Email)
	if err != nil || address.Address != req.Email {
		return errors.New("email is invalid")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, date any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(date)
	if err != nil {
		fmt.Println("json encode error:", err)
	}
}

func writeError(w http.ResponseWriter, status int, err string) {
	writeJSON(w, status, map[string]string{
		"error": err,
	})
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte("OK"))
}

func (h *Handler) UsersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		u, err := h.store.ListUsers(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, u)
	case http.MethodPost:
		h.createUserHandler(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}
}

func (h *Handler) UserHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getUserHandler(w, r)
	case http.MethodDelete:
		h.deleteHandler(w, r)
	case http.MethodPut:
		h.updateUserHandler(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}
}

func (h *Handler) getUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := helper.ParseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx := r.Context()
	u, err := h.store.FindUserByID(ctx, id)
	if err != nil {
		writeError(w, storage.StatusByError(err), err.Error())
		return
	}
	writeJSON(w, 200, u)
}

func (h *Handler) createUserHandler(w http.ResponseWriter, r *http.Request) {
	req := UserRequest{}
	err := decodeJSONBody(w, r, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	err = validateUserRequest(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	u, err := h.store.CreateUser(r.Context(), req.Name, req.Email)
	if err != nil {
		writeError(w, storage.StatusByError(err), err.Error())
		return
	}
	writeJSON(w, 201, u)
}

func (h *Handler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	id, err := helper.ParseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	err = h.store.DeleteUser(id)
	if err != nil {
		writeError(w, storage.StatusByError(err), err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{
		"status": "deleted",
	})

}

func (h *Handler) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := helper.ParseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	req := UserRequest{}
	err = decodeJSONBody(w, r, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	err = validateUserRequest(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	user, err := h.store.UpdateUser(r.Context(), id, req.Name, req.Email)
	if err != nil {
		writeError(w, storage.StatusByError(err), err.Error())
		return
	}
	writeJSON(w, 200, user)
}
