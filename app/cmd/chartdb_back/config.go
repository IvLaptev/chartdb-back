package main

import (
	"log/slog"

	"github.com/IvLaptev/chartdb-back/internal/storage/postgres"
	"github.com/IvLaptev/chartdb-back/internal/utils"
	"github.com/IvLaptev/chartdb-back/pkg/http"
	"github.com/IvLaptev/chartdb-back/pkg/s3client"
)

type Config struct {
	Logger         LoggerConfig          `yaml:"logger"`
	HTTPServer     http.HTTPServerConfig `yaml:"http_server"`
	Storage        postgres.Config       `yaml:"storage"`
	Runner         utils.RunnerConfig    `yaml:"runner"`
	S3ClientConfig s3client.S3Config     `yaml:"s3_client"`
}

type LoggerConfig struct {
	Level    slog.Level `yaml:"level"`
	Encoding string     `yaml:"encoding"`
	Path     string     `yaml:"path"`
}
