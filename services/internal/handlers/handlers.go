package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"pipecraft/internal/logger"
	"pipecraft/internal/models"
	"pipecraft/internal/services"
	"strconv"

	"github.com/gorilla/mux"
)

type PipelineService interface {
	Run(dto *models.RunPipelineRequest) (*models.RunPipelineResponse, error)
	GetStatus(id int64) (*models.PipelineStatusResponse, error)
	GetLogs(id int64) (*models.PipelineLogsResponse, error)
}

type RedisService interface {
	SetPipelineStatus(id int64, data string)
	SetPipelineLogs(id int64, data string)
	GetPipelineStatus(id int64) string
	GetPipelineLogs(id int64) string
}

type Handlers struct {
	PipelineService PipelineService
	RedisService    RedisService
}

func New(redisService RedisService, pipelineService PipelineService) *Handlers {
	return &Handlers{PipelineService: pipelineService, RedisService: redisService}
}

func (h *Handlers) RunPipeline(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	jsonData, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("error while reading json", logger.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var dto models.RunPipelineRequest
	err = json.Unmarshal(jsonData, &dto)
	if err != nil {
		slog.Error("error while parsing json", logger.Err(err))
		errorResponse := models.ErrorResponse{Error: "invalid json"}
		writeJson(errorResponse, w, http.StatusBadRequest)
		return
	}

	responseDto, err := h.PipelineService.Run(&dto)
	if err != nil {
		if errors.Is(err, services.ErrAlreadyExists) {
			w.Header().Set("Content-Type", "application/json")
			writeJson(responseDto, w, http.StatusOK)
			return
		}
		slog.Error("error while running pipeline", logger.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	writeJson(responseDto, w, http.StatusCreated) //NOTE: means that pipeline doesn't exist
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

	statusDto, err := h.PipelineService.GetStatus(pipelineId)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			errorResponseDto := models.ErrorResponse{Error: "pipeline with such id doesn't exist"}
			writeJson(errorResponseDto, w, http.StatusNotFound)
			return
		}
		slog.Error("error while getting pipeline status", logger.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//NOTE: have to make out of writeJson function because of caching
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

	logsDto, err := h.PipelineService.GetLogs(pipelineId)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			errorResponseDto := models.ErrorResponse{Error: "pipeline with such id doesn't exist or pipeline is waiting in queue"}
			writeJson(errorResponseDto, w, http.StatusNotFound)
			return
		}
		slog.Error("error while getting pipeline status", logger.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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

func writeJson(v any, w http.ResponseWriter, status int) {
	response, err := json.Marshal(v)
	if err != nil {
		slog.Error("error while marshaling json", logger.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}
