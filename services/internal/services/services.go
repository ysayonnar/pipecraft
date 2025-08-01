package services

import "pipecraft/internal/storage"

type Service struct {
	Storage *storage.Storage
}

func New(s *storage.Storage) *Service {
	return &Service{Storage: s}
}
