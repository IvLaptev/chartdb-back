package tests

import (
	"time"

	"github.com/IvLaptev/chartdb-back/internal/storage/postgres"
)

func NewPostgresTestConfig() *postgres.Config {
	return &postgres.Config{
		Host:        "localhost",
		Port:        5432,
		User:        "chartdb",
		Password:    "secretpass",
		Database:    "chartdb-test",
		SSLMode:     "disable",
		MaxIdleTime: 30 * time.Second,
		MaxLifeTime: 1 * time.Hour,
		MaxIdleConn: 1,
		MaxOpenConn: 1,
		LogLevel:    "info",
	}
}
