package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"pipecraft/internal/logger"
	"pipecraft/internal/services"
	"strconv"
)

type Handlers struct {
	PipelineService *services.PipelineService
	RedisService    *services.RedisService
}

func New(redisService *services.RedisService, pipelineService *services.PipelineService) *Handlers {
	return &Handlers{PipelineService: pipelineService, RedisService: redisService}
}

// TODO: check if pipeline exists by commit, repo, branch and image
func (h *Handlers) RunPipeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	fmt.Fprint(w, "RUN PIPELINE")
}

func (h *Handlers) PipelineStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	params := mux.Vars(r)
	strPipelineId, ok := params["id"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	pipelineId, err := strconv.ParseInt(strPipelineId, 10, 64)
	if err != nil {
		slog.Error("error while parsing pipelineId to int", logger.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cachedResponse := h.RedisService.GetPipelineStatus(pipelineId)
	if cachedResponse != "" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cachedResponse))
		return
	}

	statusDto, err := h.PipelineService.GetPipelineStatus(pipelineId)
	//TODO: handler error

	response, err := json.Marshal(statusDto)
	if err != nil {
		slog.Error("error while marshaling json", logger.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.RedisService.SetPipelineStatus(pipelineId, string(response))

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func (h *Handlers) PipelineLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	params := mux.Vars(r)
	strPipelineId, ok := params["id"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	pipelineId, err := strconv.ParseInt(strPipelineId, 10, 64)
	if err != nil {
		slog.Error("error while parsing pipelineId to int", logger.Err(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cachedResponse := h.RedisService.GetPipelineLogs(pipelineId)
	if cachedResponse != "" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cachedResponse))
		return
	}

	logsDto, err := h.PipelineService.GetPipelineLogs(pipelineId)
	//TODO: handler error

	response, err := json.Marshal(logsDto)
	if err != nil {
		slog.Error("error while marshaling json", logger.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.RedisService.SetPipelineLogs(pipelineId, string(response))

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
