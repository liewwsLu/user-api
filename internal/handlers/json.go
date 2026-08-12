package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const maxJSONBodyBytes = 1 << 20

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	var extra any
	err := decoder.Decode(&extra)
	if err == nil {
		return errors.New("request body must contain a single JSON object")
	}
	if !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
