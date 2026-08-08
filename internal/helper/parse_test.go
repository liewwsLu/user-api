package helper

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseID(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantID  int
		wantErr bool
	}{
		{name: "returns ID when valid", url: "/user?id=1", wantID: 1, wantErr: false},
		{name: "returns another valid ID", url: "/user?id=42", wantID: 42, wantErr: false},
		{name: "returns error when ID is missing", url: "/user", wantErr: true},
		{name: "returns error when id is not a number", url: "/user?id=abc", wantErr: true},
		{name: "returns error when ID is zero", url: "/user?id=0", wantErr: true},
		{name: "returns error when ID is negative", url: "/user?id=-5", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, test.url, nil)
			id, err := ParseID(req)
			if test.wantErr {
				if err == nil {
					t.Fatalf("ParseID() got: nil, want: error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseID() unexpected error: %v", err)
			}
			if id != test.wantID {
				t.Errorf("ParseID() got: %d, want: %d", id, test.wantID)
			}

		})
	}
}
