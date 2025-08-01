package handlers

import (
	"github.com/gin-gonic/gin"
	"pipecraft/internal/services"
)

type Handlers struct {
	Service *services.Service
}

func New(s *services.Service) *Handlers {
	return &Handlers{Service: s}
}

func (h *Handlers) RunPipeline(c *gin.Context) {}

func (h *Handlers) PipelineStatus(c *gin.Context) {}

func (h *Handlers) PipelineLogs(c *gin.Context) {}
