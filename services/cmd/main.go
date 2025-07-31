package main

import (
	"log/slog"
	"pipecraft/internal/app"
	"pipecraft/internal/config"
	"pipecraft/internal/logger"
)

func main() {
	cfg := config.MustParse()
	logger.BuildLogger(cfg.IsDebug)

	slog.Info("config parsed", slog.Any("config", cfg))

	application := app.New(cfg)
	application.Run()
}
