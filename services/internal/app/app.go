package app

import (
	"log/slog"
	"os"
	"os/signal"
	"pipecraft/internal/config"
	"pipecraft/internal/handlers"
	"pipecraft/internal/server"
	"pipecraft/internal/services"
	"pipecraft/internal/storage"
	"syscall"
)

type App struct {
	Config *config.Config
}

func New(cfg *config.Config) *App {
	return &App{Config: cfg}
}

func (app *App) Run() {
	storage := storage.MustInit()
	slog.Info("Database connected")

	service := services.New(storage)
	handlers := handlers.New(service)
	server := server.New(handlers)

	slog.Info("server listening", slog.Int("port", app.Config.Port))
	go server.Listen(app.Config.Port)

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	sign := <-stop

	slog.Info("Graceful shutdown application...", slog.String("signal", sign.String()))

	//TODO: проверять, есть ли не закрытые контейнеры
	storage.Db.Close()
}
