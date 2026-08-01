package config

import (
	"testing"
)

var loadTests = []struct {
	name           string
	serverPort     string
	wantServerPort int
	databaseURL    string
	wantErr        bool
}{
	{
		name:           "uses default port when port is empty",
		serverPort:     "",
		wantServerPort: 8080,
		databaseURL:    "postgres://test",
		wantErr:        false,
	},
	{
		name:           "accepts maximum valid port",
		serverPort:     "65535",
		wantServerPort: 65535,
		databaseURL:    "postgres://test",
		wantErr:        false,
	},
	{
		name:        "returns error when port is not an integer",
		serverPort:  "abc",
		databaseURL: "postgres://test",
		wantErr:     true,
	},
	{
		name:        "returns error when port is out of range",
		serverPort:  "65536",
		databaseURL: "postgres://test",
		wantErr:     true,
	},
	{
		name:        "returns error when database URL is empty",
		serverPort:  "65534",
		databaseURL: "",
		wantErr:     true,
	},
}

func TestLoad(t *testing.T) {
	for _, test := range loadTests {
		t.Run(test.name, func(t *testing.T) {
			//arrange - подготовка условий
			t.Setenv("DATABASE_URL", test.databaseURL)
			t.Setenv("SERVER_PORT", test.serverPort)
			//act
			cfg, err := Load()
			//assert
			if test.wantErr {
				if err == nil {
					t.Fatalf("Load() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}
			if cfg.ServerPort != test.wantServerPort {
				t.Errorf("ServerPort = %d, want %d", cfg.ServerPort, test.wantServerPort)
			}
			if cfg.DatabaseURL != test.databaseURL {
				t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, test.databaseURL)
			}
		})
	}
}
