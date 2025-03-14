package service_test

import (
	"errors"
	"notebook-service/api/controller"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var notebooks = []string{"notebook1", "notebook2"}

func TestGetNotebooksFailedCheckingCache(t *testing.T) {
	req := &controller.ListActiveNotebooksRequest{}

	// Mock checking cache
	errMsg := "redis error"
	redis.On("CheckCacheExists", username).Return(false, errors.New(errMsg)).Once()

	res, err := notebookService.ListActiveNotebooks(ctxWithValue, req)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, status.Error(codes.Internal, errMsg))
}

func TestGetNotebooksFailedGettingNotebookInCache(t *testing.T) {
	req := &controller.ListActiveNotebooksRequest{}

	// Mock checking cache
	redis.On("CheckCacheExists", username).Return(true, nil).Once()

	// Mock getting cache data
	errMsg := "redis error"
	redis.On("GetNotebooks", username).Return(nil, errors.New(errMsg)).Once()

	res, err := notebookService.ListActiveNotebooks(ctxWithValue, req)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, status.Error(codes.Internal, errMsg))
}

func TestGetNotebooksSuccessFromCache(t *testing.T) {
	req := &controller.ListActiveNotebooksRequest{}

	// Mock checking cache
	redis.On("CheckCacheExists", username).Return(true, nil).Once()

	// Mock getting cache data
	redis.On("GetNotebooks", username).Return(notebooks, nil).Once()

	expectedResponse := &controller.ListActiveNotebooksResponse{NotebookNames: notebooks}

	res, err := notebookService.ListActiveNotebooks(ctxWithValue, req)

	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, res)
}

func TestGetNotebooksFailedGettingNotebooksFromMongoDB(t *testing.T) {
	req := &controller.ListActiveNotebooksRequest{}

	// Mock checking cache
	redis.On("CheckCacheExists", username).Return(false, nil).Once()

	// Mock getting notebook list from mongodb
	errMsg := "mongo error"
	mongo.On("ListNotebooks", username).Return(nil, errors.New(errMsg)).Once()

	res, err := notebookService.ListActiveNotebooks(ctxWithValue, req)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, status.Error(codes.Internal, errMsg))
}

func TestGetNotebooksSuccessGettingNotebooksFromMongoDB(t *testing.T) {
	req := &controller.ListActiveNotebooksRequest{}

	// Mock checking cache
	redis.On("CheckCacheExists", username).Return(false, nil).Once()

	// Mock getting notebook list from mongodb
	mongo.On("ListNotebooks", username).Return(notebooks, nil).Once()

	// Mock caching the notebooks
	redis.On("StoreNotebooks", username, notebooks).Return(nil).Once()

	expectedResponse := &controller.ListActiveNotebooksResponse{NotebookNames: notebooks}

	res, err := notebookService.ListActiveNotebooks(ctxWithValue, req)

	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, res)
}
