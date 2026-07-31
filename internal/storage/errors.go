package storage

import (
	"errors"
	"net/http"
)

var ErrValidation = errors.New("validation error")
var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("already exist")

func StatusByError(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if errors.Is(err, ErrValidation) {
		return http.StatusBadRequest
	}
	if errors.Is(err, ErrNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, ErrConflict) {
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}
