package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"pipecraft/internal/config"
	"pipecraft/internal/handlers"
	"pipecraft/internal/logger"
	"time"
)

type Server struct {
	Handlers *handlers.Handlers
}

func New(h *handlers.Handlers) *Server {
	return &Server{Handlers: h}
}

func (s *Server) Listen(httpCfg config.Http) {
	r := mux.NewRouter()

	r.HandleFunc("/run-pipeline", s.Handlers.RunPipeline)
	r.HandleFunc("/pipeline/{id}/status", s.Handlers.PipelineStatus)
	r.HandleFunc("/pipeline/{id}/logs", s.Handlers.PipelineLogs)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", httpCfg.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(httpCfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(httpCfg.WriteTimeout) * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		slog.Error("error while listening http-server", logger.Err(err))
	}
}
