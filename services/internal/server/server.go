package server

import (
	"fmt"
	"log/slog"
	"pipecraft/internal/handlers"
	"pipecraft/internal/logger"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Handlers *handlers.Handlers
}

func New(h *handlers.Handlers) *Server {
	return &Server{Handlers: h}
}

func (s *Server) Listen(port int) {
	r := gin.New()
	r.Use(gin.Recovery()) // handles panics and sends back status 500

	r.POST("/run-pipeline", s.Handlers.RunPipeline)
	r.GET("/pipeline/:id/status", s.Handlers.PipelineStatus)
	r.GET("/pipeline/:id/logs", s.Handlers.PipelineLogs)

	err := r.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		slog.Error("error while running server", logger.Err(err))
		return
	}
}
