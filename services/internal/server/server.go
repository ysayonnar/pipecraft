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
	r := gin.Default()

	err := r.Run(fmt.Sprintf(":%d", port))
	if err != nil {
		slog.Error("error while running server", logger.Err(err))
		return
	}
}
