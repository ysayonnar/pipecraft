package main

import (
	"log/slog"
	"pipecraft/internal/config"
	"pipecraft/internal/logger"
)

const DEBUG = true

func main() {
	logger.BuildLogger(DEBUG)

	cfg := config.MustParse()
	slog.Info("config parsed", slog.Any("config", cfg))

	for {
	}
}
