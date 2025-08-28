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
	"pipecraft/internal/worker"
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

	pipelineService := services.NewPipelineService(storage)
	redisService := services.NewRedisService()
	slog.Info("redis connected")

	handlers := handlers.New(redisService, pipelineService)
	server := server.New(handlers)

	slog.Info("server listening", slog.Int("port", app.Config.Http.Port))
	go server.Listen(app.Config.Http)

	go worker.StartListener(storage)

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	sign := <-stop

	slog.Info("Graceful shutdown application...", slog.String("signal", sign.String()))

	redisService.Close()
	storage.Db.Close()
}
