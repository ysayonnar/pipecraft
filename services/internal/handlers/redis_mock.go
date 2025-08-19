package handlers

import (
	"fmt"

	"github.com/alicebob/miniredis/v2"
)

const DEFAULT_TTL_SECONDS = 30

type MockRedisService struct {
	client *miniredis.Miniredis
}

func NewMockRedisServie() *MockRedisService {
	client, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	return &MockRedisService{client: client}
}

func (m MockRedisService) SetPipelineStatus(id int64, data string) {
	key := fmt.Sprintf("status:%d", id)

	if err := m.client.Set(key, data); err != nil {
		panic(err)
	}

	m.client.SetTTL(key, DEFAULT_TTL_SECONDS)
}

func (m MockRedisService) SetPipelineLogs(id int64, data string) {
	key := fmt.Sprintf("logs:%d", id)

	if err := m.client.Set(key, data); err != nil {
		panic(err)
	}

	m.client.SetTTL(key, DEFAULT_TTL_SECONDS)
}

func (m MockRedisService) GetPipelineStatus(id int64) string {
	key, err := m.client.Get(fmt.Sprintf("status:%d", id))
	if err != nil {
		panic(err)
	}

	return key
}

func (m MockRedisService) GetPipelineLogs(id int64) string {
	key, err := m.client.Get(fmt.Sprintf("logs:%d", id))
	if err != nil {
		panic(err)
	}

	return key
}
