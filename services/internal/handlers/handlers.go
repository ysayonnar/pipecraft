package handlers

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"pipecraft/internal/services"
)

type Handlers struct {
	Service *services.Service
}

func New(s *services.Service) *Handlers {
	return &Handlers{Service: s}
}

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

	pipelineId, ok := params["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "STATUS pipeline_id: %s", pipelineId)
}

func (h *Handlers) PipelineLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	params := mux.Vars(r)

	pipelineId, ok := params["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "LOGS pipeline_id: %s", pipelineId)
}
