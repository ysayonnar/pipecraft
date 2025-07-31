package app

import (
	"log/slog"
	"pipecraft/internal/config"
	"pipecraft/internal/storage"
)

type App struct {
	Config *config.Config
}

func New(cfg *config.Config) *App {
	return &App{Config: cfg}
}

func (app *App) Run() {
	s := storage.MustInit()

	_ = s

	slog.Info("Database connected")

	// TODO: инициализировать сервисы для БИЗНЕС-ЛОГИКИ!!!

	// TODO: иницилизировать хендлеры

	// TODO: инициализировать и поднять сервер
}
