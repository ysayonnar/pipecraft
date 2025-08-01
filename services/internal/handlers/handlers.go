package handlers

import (
	"pipecraft/internal/services"
)

type Handlers struct {
	Service *services.Service
}

func New(s *services.Service) *Handlers {
	return &Handlers{Service: s}
}
