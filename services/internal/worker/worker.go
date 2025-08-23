package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"pipecraft/internal/logger"
	"pipecraft/internal/storage"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

const (
	MAX_WORKERS     = 5
	LISTEN_INTERVAL = 10
)

type Worker struct {
	DockerClient *client.Client
	PipelineId   int64
	done         chan bool
}

func StartListener(s *storage.Storage) {
	workerPool := make(chan struct{}, MAX_WORKERS)

	for {
		pipelineId, err := s.GetLastWaitingPipeline()
		if err != nil {
			if !errors.Is(err, storage.ErrNotFound) {
				slog.Error("error while getting last pipeline waiting pipeline", logger.Err(err))
			}
		}

		workerPool <- struct{}{}
		go func(pipelineId int64) {
			defer func() { <-workerPool }()
			worker := NewWorker(pipelineId)
			<-worker.done
		}(pipelineId)

		time.Sleep(time.Duration(LISTEN_INTERVAL) * time.Second)
	}
}

func NewWorker(pipelineId int64) *Worker {
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		slog.Error("error while creating docker client", logger.Err(err))
		panic(err)
	}
	return &Worker{PipelineId: pipelineId, done: make(chan bool), DockerClient: client}
}

func (w *Worker) Run() {
	const op = "worker.Run"

	containerName := fmt.Sprintf("pipecraft-%d", w.PipelineId)

	ctx := context.Background()
	resp, err := w.DockerClient.ContainerCreate(
		ctx,
		&container.Config{
			Image: "alpine-git",
		},
		&container.HostConfig{
			Binds: []string{
				"/var/run/docker.sock:/var/run/docker.sock",
			},
		},
		nil,
		nil,
		containerName,
	)
	if err != nil {
		//TODO: менять статус пайплайна
		slog.Error("error while creating docker container", logger.Err(err))
	}

	_ = resp

	// TODO: клонировать репозиторий

	// TODO: парсить jobs для CI

	// TODO: выполнение jobs и запись логов

	// TODO: изменение статуса пайплайна

	// TODO: очистка ресурсов и пишу в w.done
}
