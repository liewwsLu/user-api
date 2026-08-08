package helper

import (
	"errors"
	"net/http"
	"strconv"
)

func ParseID(r *http.Request) (int, error) {
	idT := r.URL.Query().Get("id")
	if idT == "" {
		return 0, errors.New("invalid id")
	}
	id, err := strconv.Atoi(idT)
	if err != nil {
		return 0, errors.New("invalid id")
	}
	if id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}
