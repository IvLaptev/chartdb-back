package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/IvLaptev/chartdb-back/internal/utils"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/spf13/pflag"
)

func main() {
	ctx := context.Background()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := setupLogger(cfg.Logger)

	logger.Info("application started")
	app := newApplication(cfg, logger)
	if err = app.Run(ctx); err != nil && !errors.Is(err, utils.ErrSignalExit) {
		logger.Error("application stopped with error", slog.Any("error", err))
	} else {
		logger.Info("application stopped")
	}
}

func loadConfig() (*Config, error) {
	var config Config

	configPath := pflag.StringP("config", "c", "", "config path")
	pflag.Parse()

	err := cleanenv.ReadConfig(*configPath, &config)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	return &config, nil
}

func setupLogger(conf LoggerConfig) *slog.Logger {
	var logger *slog.Logger
	var output io.Writer = os.Stdout

	switch conf.Encoding {
	case "console":
		logger = slog.New(slog.NewTextHandler(output, &slog.HandlerOptions{
			Level: conf.Level,
		}))
	case "json":
		logger = slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{
			Level: conf.Level,
		}))
	}

	return logger
}
