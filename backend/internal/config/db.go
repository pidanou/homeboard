package config

import (
	"os"
	"strconv"
)

type DBConfig struct {
	MaxConns int32
	MinConns int32
}

// LoadDB reads database connection pool sizing from the environment. pgx's own
// default (max(4, NumCPU)) is too small for a web API where a single page load
// fires several parallel requests, so we default higher and let it be tuned
// via env vars instead of DATABASE_URL query params (which golang-migrate's
// driver would forward to Postgres verbatim and fail on).
func LoadDB() DBConfig {
	cfg := DBConfig{MaxConns: 10, MinConns: 2}
	if v, err := strconv.Atoi(os.Getenv("DB_POOL_MAX_CONNS")); err == nil {
		cfg.MaxConns = int32(v)
	}
	if v, err := strconv.Atoi(os.Getenv("DB_POOL_MIN_CONNS")); err == nil {
		cfg.MinConns = int32(v)
	}
	return cfg
}
