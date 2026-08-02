package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "returns OK for GET",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
			wantBody:   "OK",
		},
		{
			name:       "returns method not allowed for POST",
			method:     http.MethodPost,
			wantStatus: http.StatusMethodNotAllowed,
			wantBody:   "method not allowed\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := New(nil)
			request := httptest.NewRequest(test.method, "/health", nil)
			recorder := httptest.NewRecorder()
			h.HealthHandler(recorder, request)
			if recorder.Code != test.wantStatus {
				t.Errorf("HealthHandler() status = %d, want = %d",
					recorder.Code,
					test.wantStatus)
			}
			if recorder.Body.String() != test.wantBody {
				t.Errorf("HealthHandler() body = %q, want = %q",
					recorder.Body.String(),
					test.wantBody)
			}
		})
	}
}
