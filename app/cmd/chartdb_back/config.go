package main

import (
	"log/slog"

	"github.com/IvLaptev/chartdb-back/internal/storage/postgres"
	"github.com/IvLaptev/chartdb-back/pkg/emailsender"
	"github.com/IvLaptev/chartdb-back/pkg/http"
	"github.com/IvLaptev/chartdb-back/pkg/s3client"
	"github.com/IvLaptev/chartdb-back/pkg/utils"
)

type Config struct {
	Logger         LoggerConfig                  `yaml:"logger"`
	HTTPServer     http.HTTPServerConfig         `yaml:"http_server"`
	Storage        postgres.Config               `yaml:"storage"`
	Runner         utils.RunnerConfig            `yaml:"runner"`
	S3ClientConfig s3client.S3Config             `yaml:"s3_client"`
	EmailSender    emailsender.EmailSenderConfig `yaml:"email_sender"`
	Auth           AuthConfig                    `yaml:"auth"`
}

type LoggerConfig struct {
	Level    slog.Level `yaml:"level"`
	Encoding string     `yaml:"encoding"`
	Path     string     `yaml:"path"`
}

type AuthConfig struct {
	TokenSecret string `yaml:"token_secret" env:"AUTH_TOKEN_SECRET"`
}
