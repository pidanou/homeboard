package config

import "testing"

func TestLoadDBDefaults(t *testing.T) {
	t.Setenv("DB_POOL_MAX_CONNS", "")
	t.Setenv("DB_POOL_MIN_CONNS", "")

	cfg := LoadDB()
	if cfg.MaxConns != 10 {
		t.Errorf("want default MaxConns 10, got %d", cfg.MaxConns)
	}
	if cfg.MinConns != 2 {
		t.Errorf("want default MinConns 2, got %d", cfg.MinConns)
	}
}

func TestLoadDBOverrides(t *testing.T) {
	t.Setenv("DB_POOL_MAX_CONNS", "25")
	t.Setenv("DB_POOL_MIN_CONNS", "5")

	cfg := LoadDB()
	if cfg.MaxConns != 25 {
		t.Errorf("want MaxConns 25, got %d", cfg.MaxConns)
	}
	if cfg.MinConns != 5 {
		t.Errorf("want MinConns 5, got %d", cfg.MinConns)
	}
}

func TestLoadDBInvalidFallsBackToDefault(t *testing.T) {
	t.Setenv("DB_POOL_MAX_CONNS", "not-a-number")
	t.Setenv("DB_POOL_MIN_CONNS", "")

	cfg := LoadDB()
	if cfg.MaxConns != 10 {
		t.Errorf("want default MaxConns 10 on invalid input, got %d", cfg.MaxConns)
	}
}
