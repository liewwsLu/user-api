package helper

import (
	"errors"
	"net/http"
	"strconv"
)

func ParseID(r *http.Request) (int, error) {
	idText := r.URL.Query().Get("id")
	if idText == "" {
		return 0, errors.New("invalid id")
	}
	id, err := strconv.Atoi(idText)
	if err != nil {
		return 0, errors.New("invalid id")
	}
	if id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}
