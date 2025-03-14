package mock_redis

import (
	"github.com/stretchr/testify/mock"
)

// MockRedisRepo is mocking repository layer of redis
type MockRedis struct {
	mock.Mock
}

func (r *MockRedis) CheckCacheExists(username string) (bool, error) {
	args := r.Called(username)
	return args.Bool(0), args.Error(1)
}

func (r *MockRedis) StoreNotebooks(username string, notebookNames []string) error {
	args := r.Called(username, notebookNames)
	return args.Error(0)
}

func (r *MockRedis) AddNotebook(username string, notebookName string) error {
	args := r.Called(username, notebookName)
	return args.Error(0)
}

func (r *MockRedis) GetNotebooks(username string) ([]string, error) {
	args := r.Called(username)

	if notebooks, ok := args.Get(0).([]string); ok {
		return notebooks, args.Error(1)
	}

	return nil, args.Error(1)
}

func (r *MockRedis) DeleteNotebook(username string, notebookName string) error {
	args := r.Called(username, notebookName)
	return args.Error(0)
}
