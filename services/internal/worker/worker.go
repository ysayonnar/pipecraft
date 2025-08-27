package worker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"pipecraft/internal/jobs"
	"pipecraft/internal/logger"
	"pipecraft/internal/storage"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

const (
	MAX_WORKERS     = 5
	LISTEN_INTERVAL = 10

	DEFAULT_CI_CONFIG_PATH = "/workspace/ci.yaml"
)

type Worker struct {
	dockerClient *client.Client
	storage      *storage.Storage
	pipelineId   int64
	done         chan bool
}

func StartListener(s *storage.Storage) {
	workerPool := make(chan struct{}, MAX_WORKERS)

	// TODO: почему то запускается несколько воркеров

	for {
		pipelineId, err := s.GetLastWaitingPipeline()
		if err != nil {
			if !errors.Is(err, storage.ErrNotFound) {
				slog.Error("error while getting last pipeline waiting pipeline", logger.Err(err))
			}
		}

		if pipelineId != 0 {
			go func(pipelineId int64) {
				workerPool <- struct{}{}

				worker := NewWorker(s, pipelineId)
				err := worker.storage.UpdatePipelineStatus(pipelineId, storage.PIPELINE_STATUS_RUNNING)
				if err != nil {
					slog.Warn("pipeline with known id was not found")
				} else {
					go worker.Run()
				}

				<-worker.done
				<-workerPool
			}(pipelineId)
		}

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

	slog.Debug("running pipeline", slog.Int64("pipeline_id", w.pipelineId))

	defer func() { w.done <- true }()

	//TODO: нужен dind
	ctx := context.Background()
	resp, err := w.dockerClient.ContainerCreate(
		ctx,
		&container.Config{
			Image:      "alpine-git",
			WorkingDir: "/workspace",
			Cmd:        []string{"sleep", "infinity"},
		},
		&container.HostConfig{
			Binds: []string{
				"/var/run/docker.sock:/var/run/docker.sock",
			},
		},
		nil,
		nil,
		fmt.Sprintf("pipeline-%d", w.pipelineId),
	)
	if err != nil {
		slog.Error("error while creating docker container", logger.Err(err))
		err := w.storage.UpdatePipelineStatus(w.pipelineId, storage.PIPELINE_STATUS_ABORTED)
		if err != nil {
			slog.Error("error while updating pipeline status", logger.Err(err))
		}
		return
	}

	if err := w.dockerClient.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		slog.Error("error while starting container docker container", logger.Err(err))
		err := w.storage.UpdatePipelineStatus(w.pipelineId, storage.PIPELINE_STATUS_ABORTED)
		if err != nil {
			slog.Error("error while updating pipeline status", logger.Err(err))
		}
		return
	}

	defer func() {
		err = w.cleanupContainer(resp.ID)
		if err != nil {
			slog.Warn("failed to stop container", logger.Err(err))
		}
	}()

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

	// reading ci config file
	jobs, err := w.readCiConfig(resp.ID)
	if err != nil {
		slog.Error("error while reading CI config", logger.Err(err))
		err := w.storage.UpdatePipelineStatus(w.pipelineId, storage.PIPELINE_STATUS_ABORTED)
		if err != nil {
			slog.Error("error while updating pipeline status", logger.Err(err))
		}
		return
	}

	slog.Debug("jobs", jobs)

	for jobNumber, job := range jobs {
		for _, step := range job.Steps {
			execConfig := container.ExecOptions{
				Cmd:          strings.Split(step.Run, " "),
				AttachStdout: true,
				AttachStderr: true,
			}

			logs, exitCode, err := w.execCommandWithLogs(resp.ID, execConfig)
			if err != nil {
				slog.Error("error while executing job step", logger.Err(err))
				err := w.storage.UpdatePipelineStatus(w.pipelineId, storage.PIPELINE_STATUS_ABORTED)
				if err != nil {
					slog.Error("error while updating pipeline status", logger.Err(err))
				}
				return
			}
			if exitCode != 0 {
				err := w.storage.UpdatePipelineStatus(w.pipelineId, storage.PIPELINE_STATUS_FAILED)
				if err != nil {
					slog.Error("error while updating pipeline status", logger.Err(err))
				}

				err = w.storage.CreateLog(storage.LogsTable{
					CommandNumber: jobNumber,
					CommandName:   fmt.Sprintf("%s:%s", job.Name, step.Name),
					Command:       step.Run,
					Results:       string(logs),
					FinalStatus:   fmt.Sprintf("Failed, exit code: %d", exitCode),
					PipelineId:    w.pipelineId,
				})
				if err != nil {
					slog.Error("error while creating logs", logger.Err(err))
				}

				return
			}

			err = w.storage.CreateLog(storage.LogsTable{
				CommandNumber: jobNumber,
				CommandName:   fmt.Sprintf("%s:%s", job.Name, step.Name),
				Command:       step.Run,
				Results:       string(logs),
				FinalStatus:   "Succeeded",
				PipelineId:    w.pipelineId,
			})
			if err != nil {
				slog.Error("error while creating logs", logger.Err(err))
				return
			}
		}
	}

	err = w.storage.UpdatePipelineStatus(w.pipelineId, storage.PIPELINE_STATUS_COMPLETED)
	if err != nil {
		slog.Error("error while updating pipeline status", logger.Err(err))
		return
	}

	// TODO: c вольюмами что то придумать чтобы cd работал
}

func (w *Worker) cloneRepository(containerId, repository, branch, commit string) error {
	const op = `worker.cloneRepository`

	execConfig1 := container.ExecOptions{
		Cmd: []string{"git", "clone", "--depth", "1", "--branch", branch, "--single-branch", repository, "/workspace"},
	}

	execConfig2 := container.ExecOptions{
		Cmd: []string{"git", "-C", "/workspace", "checkout", commit},
	}

	logs, exitCode, err := w.execCommandWithLogs(containerId, execConfig1)
	if err != nil || exitCode != 0 {
		slog.Error("error while cloning repository", logger.Err(err), slog.String("logs", string(logs)), slog.Int("exitCode", exitCode))
		return fmt.Errorf("op: %s, err: %w", op, err)
	}

	logs, exitCode, err = w.execCommandWithLogs(containerId, execConfig2)
	if err != nil || exitCode != 0 {
		slog.Error("error while cloning repository", logger.Err(err), slog.String("logs", string(logs)), slog.Int("exitCode", exitCode))
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

func (w *Worker) execCommandWithLogs(containerId string, execOpts container.ExecOptions) ([]byte, int, error) {
	const op = `worker.ExecCommandWithLogs`

	ctx := context.Background()

	execIDResp, err := w.dockerClient.ContainerExecCreate(ctx, containerId, execOpts)
	if err != nil {
		return nil, 0, fmt.Errorf("op: %s, err: %w", op, err)
	}

	attachResp, err := w.dockerClient.ContainerExecAttach(ctx, execIDResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("op: %s, err: %w", op, err)
	}
	defer attachResp.Close()

	var exitCode int
	for {
		inspect, err := w.dockerClient.ContainerExecInspect(ctx, execIDResp.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("op: %s, err: %w", op, err)
		}
		if !inspect.Running {
			exitCode = inspect.ExitCode
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	var outBuf, errBuf bytes.Buffer
	_, err = stdcopy.StdCopy(&outBuf, &errBuf, attachResp.Reader)
	if err != nil {
		return nil, 0, fmt.Errorf("op: %s, err: %w", op, err)
	}

	return outBuf.Bytes(), exitCode, nil
}

func (w *Worker) readCiConfig(containerId string) ([]jobs.Job, error) {
	const op = `worker.readCiConfig`
	execConfig := container.ExecOptions{
		Cmd:          []string{"cat", DEFAULT_CI_CONFIG_PATH},
		AttachStdout: true,
		AttachStderr: true,
	}

	data, _, err := w.execCommandWithLogs(containerId, execConfig)
	if err != nil {
		return nil, fmt.Errorf("op: %s, err: %w", op, err)
	}

	jobs, err := jobs.ParseJobsOrdered(data)
	if err != nil {
		return nil, fmt.Errorf("op: %s, err: %w", op, err)
	}

	return jobs, nil
}

func (w *Worker) cleanupContainer(containerID string) error {
	const op = `worker.cleanupContainer`

	ctx := context.Background()

	if err := w.dockerClient.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		return fmt.Errorf("op: %s, err: %w", op, err)
	}

	if err := w.dockerClient.ContainerRemove(ctx, containerID, container.RemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}); err != nil {
		return fmt.Errorf("op: %s, err: %w", op, err)
	}

	return nil
}
