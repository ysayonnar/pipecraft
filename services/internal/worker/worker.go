package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"pipecraft/internal/logger"
	"pipecraft/internal/storage"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

const (
	MAX_WORKERS     = 5
	LISTEN_INTERVAL = 10
)

type Worker struct {
	dockerClient *client.Client
	storage      *storage.Storage
	pipelineId   int64
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
			worker := NewWorker(s, pipelineId)
			<-worker.done
		}(pipelineId)

		time.Sleep(time.Duration(LISTEN_INTERVAL) * time.Second)
	}
}

func NewWorker(s *storage.Storage, pipelineId int64) *Worker {
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		slog.Error("error while creating docker client", logger.Err(err))
		panic(err)
	}
	return &Worker{storage: s, pipelineId: pipelineId, done: make(chan bool), dockerClient: client}
}

func (w *Worker) Run() {
	const op = "worker.Run"

	defer func() { w.done <- true }()

	err := w.storage.UpdatePipelineStatus(w.pipelineId, storage.PIPELINE_STATUS_RUNNING)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			slog.Warn("pipeline with known id was not found")
			return
		}
		slog.Error("error while updating pipeline status", logger.Err(err))
		return
	}

	ctx := context.Background()
	resp, err := w.dockerClient.ContainerCreate(
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
		fmt.Sprintf("pipecraft-%d", w.pipelineId),
	)
	if err != nil {
		slog.Error("error while creating docker container", logger.Err(err))
		err := w.storage.UpdatePipelineStatus(w.pipelineId, storage.PIPELINE_STATUS_ABORTED)
		if err != nil {
			slog.Error("error while updating pipeline status", logger.Err(err))
		}
		return
	}

	pipelineInfo, err := w.storage.GetPipelineInfo(w.pipelineId)
	if err != nil {
		slog.Error("error while selecting pipeline info", logger.Err(err))
		err := w.storage.UpdatePipelineStatus(w.pipelineId, storage.PIPELINE_STATUS_ABORTED)
		if err != nil {
			slog.Error("error while updating pipeline status", logger.Err(err))
		}
		return
	}

	// cloning repository
	err = w.cloneRepository(resp.ID, pipelineInfo.Repository, pipelineInfo.Branch, pipelineInfo.Commit)
	if err != nil {
		slog.Error("error while cloning repository", logger.Err(err))
		err := w.storage.UpdatePipelineStatus(w.pipelineId, storage.PIPELINE_STATUS_ABORTED)
		if err != nil {
			slog.Error("error while updating pipeline status", logger.Err(err))
		}
		return
	}

	// TODO: парсить jobs для CI

	// TODO: выполнение jobs и запись логов

	// TODO: изменение статуса пайплайна

	// TODO: очистка ресурсов и пишу в w.done
}

func (w *Worker) cloneRepository(containerId, repository, branch, commit string) error {
	const op = `worker.cloneRepository`

	execConfig1 := container.ExecOptions{
		Cmd: []string{"git", "clone", "--depth", "1", "--branch", branch, "--single-branch", repository, "/workspace"},
	}

	execConfig2 := container.ExecOptions{
		Cmd: []string{"git", "-C", "/workspace", "checkout", commit},
	}

	exitCode, err := w.execCommandWithExitCode(containerId, execConfig1)
	if err != nil || exitCode != 0 {
		return fmt.Errorf("op: %s, err: %w", op, err)
	}

	exitCode, err = w.execCommandWithExitCode(containerId, execConfig2)
	if err != nil || exitCode != 0 {
		return fmt.Errorf("op: %s, err: %w", op, err)
	}

	return nil
}

func (w *Worker) execCommandWithExitCode(containerId string, execOpts container.ExecOptions) (int, error) {
	const op = `worker.ExecCommandWithExitCode`

	ctx := context.Background()

	execIDResp, err := w.dockerClient.ContainerExecCreate(ctx, containerId, execOpts)
	if err != nil {
		return 0, fmt.Errorf("op: %s, err: %w", op, err)
	}

	if err = w.dockerClient.ContainerExecStart(ctx, execIDResp.ID, container.ExecAttachOptions{}); err != nil {
		return 0, fmt.Errorf("op: %s, err: %w", op, err)
	}

	inspectResp, err := w.dockerClient.ContainerExecInspect(ctx, execIDResp.ID)
	if err != nil {
		return 0, fmt.Errorf("op: %s, err: %w", op, err)
	}

	return inspectResp.ExitCode, nil
}

func (w *Worker) execCommandWithLogs(containerId string, execOpts container.ExecOptions) (*types.HijackedResponse, error) {
	const op = `worker.ExecCommandWithLogs`

	ctx := context.Background()

	execIDResp, err := w.dockerClient.ContainerExecCreate(ctx, containerId, execOpts)
	if err != nil {
		return nil, fmt.Errorf("op: %s, err: %w", op, err)
	}

	attachResp, err := w.dockerClient.ContainerExecAttach(ctx, execIDResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return nil, fmt.Errorf("op: %s, err: %w", op, err)
	}

	return &attachResp, nil
}
