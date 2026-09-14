package config

import (
	"os"
	"reflect"
	"testing"
)

func TestGetEnvListSplit(t *testing.T) {
	cases := []struct {
		value string
		want  []string
	}{
		{"http://a.com,http://b.com", []string{"http://a.com", "http://b.com"}},
		{"http://a.com, http://b.com ", []string{"http://a.com", "http://b.com"}},
		{"http://only.com", []string{"http://only.com"}},
	}
	for _, c := range cases {
		t.Setenv("CORS_LIST", c.value)
		got := getEnvList("CORS_LIST", "")
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("getEnvList(%q) = %v, want %v", c.value, got, c.want)
		}
	}
}

func TestGetEnvListUnsetUsesFallback(t *testing.T) {
	t.Setenv("CORS_LIST_UNSET", "")
	got := getEnvList("CORS_LIST_UNSET", "http://localhost:8080")
	if !reflect.DeepEqual(got, []string{"http://localhost:8080"}) {
		t.Errorf("fallback = %v, want [http://localhost:8080]", got)
	}
}

func TestConfigLoadDefaults(t *testing.T) {
	cfg := Load()
	if cfg.CORSAllowedOrigins == nil || len(cfg.CORSAllowedOrigins) == 0 {
		t.Fatal("expected default CORS origins")
	}
	if cfg.CORSAllowedOrigins[0] != "http://localhost:8080" {
		t.Errorf("default CORS origin = %v, want http://localhost:8080", cfg.CORSAllowedOrigins[0])
	}
	if cfg.ServerPort != "8080" {
		t.Errorf("ServerPort = %q, want 8080", cfg.ServerPort)
	}
}

func TestConfigLoadUsesEnv(t *testing.T) {
	old := os.Getenv("DB_HOST")
	os.Setenv("DB_HOST", "db.example.com")
	defer func() {
		if old == "" {
			os.Unsetenv("DB_HOST")
		} else {
			os.Setenv("DB_HOST", old)
		}
	}()
	cfg := Load()
	if cfg.DatabaseURL == "" {
		t.Fatal("expected DatabaseURL to be built")
	}
}
