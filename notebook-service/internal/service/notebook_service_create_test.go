package service_test

import (
	"context"
	"errors"
	"log"
	"notebook-service/api/controller"
	"notebook-service/internal/auth"
	"notebook-service/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/rest"
)

func TestCreateNotebookInvalidRequests(t *testing.T) {
	ctx := context.Background()

	// Invalid max cpu
	req := &controller.CreateNotebookRequest{
		Name:   "notebook-test",
		MaxCpu: stringPtr("a"),
	}

	_, err := notebookService.CreateNotebook(ctx, req)

	assert.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid max cpu"))

	// Invalid min cpu
	req.MaxCpu = stringPtr("2")
	req.MinCpu = stringPtr("a")

	_, err = notebookService.CreateNotebook(ctx, req)

	assert.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid min cpu"))

	// Invalid max memory
	req.MinCpu = stringPtr("1")
	req.MaxMemory = stringPtr("a")

	_, err = notebookService.CreateNotebook(ctx, req)

	assert.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid max memory"))

	// Invalid min memory
	req.MaxMemory = stringPtr("2G")
	req.MinMemory = stringPtr("a")

	_, err = notebookService.CreateNotebook(ctx, req)

	assert.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid min memory"))

	// Invalid volume size
	req.MinMemory = stringPtr("1G")
	req.Volume = stringPtr("a")

	_, err = notebookService.CreateNotebook(ctx, req)

	assert.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid volume size"))
}

func TestCreateNotebookGetKubeConfigError(t *testing.T) {
	ctx := context.Background()

	req := &controller.CreateNotebookRequest{
		Name: "notebook-test",
	}

	restoreGetKubeConfig := mockGetKubeConfig(nil, errors.New("get kube config error"))
	defer restoreGetKubeConfig()

	res, err := notebookService.CreateNotebook(ctx, req)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, status.Error(codes.Internal, "failed getting kube config"))
}

func TestCreateNotebookErrorCreateDynamicClient(t *testing.T) {
	ctx := context.Background()

	req := &controller.CreateNotebookRequest{
		Name: "notebook-test",
	}

	restoreGetKubeConfig := mockGetKubeConfig(&rest.Config{}, nil)
	defer restoreGetKubeConfig()

	restoreCreateDynamicClient := mockCreateDynamicClient(errors.New("error at creating dynamic client"))
	defer restoreCreateDynamicClient()

	res, err := notebookService.CreateNotebook(ctx, req)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, status.Error(codes.Internal, "failed creating dynamic client"))
}

func TestCreateNotebookErrorFailedGeneratingPVC(t *testing.T) {
	ctx := context.Background()

	req := &controller.CreateNotebookRequest{
		Name: "notebook-test",
	}

	restoreGetKubeConfig := mockGetKubeConfig(&rest.Config{}, nil)
	defer restoreGetKubeConfig()

	restoreCreateDynamicClient := mockCreateDynamicClient(nil)
	defer restoreCreateDynamicClient()

	restoreCreatePVCResource := mockCreatePvcResource(nil, errors.New("pvc error"))
	defer restoreCreatePVCResource()

	res, err := notebookService.CreateNotebook(ctx, req)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, status.Error(codes.Internal, "failed creating pvc"))
}

func TestCreateNotebookFailedStoringNotebookToMongoDB(t *testing.T) {
	req := &controller.CreateNotebookRequest{
		Name: "notebook-test",
	}

	restoreGetKubeConfig := mockGetKubeConfig(&rest.Config{}, nil)
	defer restoreGetKubeConfig()

	restoreCreateDynamicClient := mockCreateDynamicClient(nil)
	defer restoreCreateDynamicClient()

	restoreCreatePVCResource := mockCreatePvcResource(pvc, nil)
	defer restoreCreatePVCResource()

	notebook := &model.NotebookEntity{
		Username:     ctxWithValue.Value(auth.CtxKey).(string),
		NotebookName: req.Name,
	}

	errMsg := "mongo error"
	mongo.On("CreateNotebook", notebook).Return(errors.New(errMsg)).Once()

	res, err := notebookService.CreateNotebook(ctxWithValue, req)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, status.Error(codes.Internal, errMsg))
}

func TestCreateNotebookFailedCheckingCache(t *testing.T) {
	req := &controller.CreateNotebookRequest{
		Name: "notebook-test",
	}

	restoreGetKubeConfig := mockGetKubeConfig(&rest.Config{}, nil)
	defer restoreGetKubeConfig()

	restoreCreateDynamicClient := mockCreateDynamicClient(nil)
	defer restoreCreateDynamicClient()

	restoreCreatePVCResource := mockCreatePvcResource(pvc, nil)
	defer restoreCreatePVCResource()

	restoreCallOpen := mockCallOpen()
	defer restoreCallOpen()

	notebook := &model.NotebookEntity{
		Username:     ctxWithValue.Value(auth.CtxKey).(string),
		NotebookName: req.Name,
	}

	// Mock storing notebook
	mongo.On("CreateNotebook", notebook).Return(nil).Once()

	// Mock checking the cache
	errMsg := "redis error"
	redis.On("CheckCacheExists", notebook.Username).Return(false, errors.New(errMsg)).Once()

	res, err := notebookService.CreateNotebook(ctxWithValue, req)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, status.Error(codes.Internal, errMsg))
}

func TestCreateNotebookFailedCachingNotebook(t *testing.T) {
	req := &controller.CreateNotebookRequest{
		Name: "notebook-test",
	}

	restoreGetKubeConfig := mockGetKubeConfig(&rest.Config{}, nil)
	defer restoreGetKubeConfig()

	restoreCreateDynamicClient := mockCreateDynamicClient(nil)
	defer restoreCreateDynamicClient()

	restoreCreatePVCResource := mockCreatePvcResource(pvc, nil)
	defer restoreCreatePVCResource()

	notebook := &model.NotebookEntity{
		Username:     ctxWithValue.Value(auth.CtxKey).(string),
		NotebookName: req.Name,
	}

	// Mock storing notebook
	mongo.On("CreateNotebook", notebook).Return(nil).Once()

	// Mock checking the cache
	redis.On("CheckCacheExists", notebook.Username).Return(true, nil).Once()

	// Mock failed adding notebook to existing cache
	errMsg := "redis error"
	redis.On("AddNotebook", notebook.Username, req.Name).Return(errors.New(errMsg)).Once()

	res, err := notebookService.CreateNotebook(ctxWithValue, req)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, status.Error(codes.Internal, errMsg))
}

func TestCreateNotebookSuccess(t *testing.T) {
	req := &controller.CreateNotebookRequest{
		Name: "notebook-test",
	}

	restoreGetKubeConfig := mockGetKubeConfig(&rest.Config{}, nil)
	defer restoreGetKubeConfig()

	restoreCreateDynamicClient := mockCreateDynamicClient(nil)
	defer restoreCreateDynamicClient()

	restoreCreatePVCResource := mockCreatePvcResource(pvc, nil)
	defer restoreCreatePVCResource()

	notebook := &model.NotebookEntity{
		Username:     ctxWithValue.Value(auth.CtxKey).(string),
		NotebookName: req.Name,
	}

	// Mock storing notebook
	mongo.On("CreateNotebook", notebook).Return(nil).Once()

	// Mock checking the cache
	redis.On("CheckCacheExists", notebook.Username).Return(false, nil).Once()

	_, err := notebookService.CreateNotebook(ctxWithValue, req)
	log.Println(err)
	assert.Nil(t, err)
}

// Helper functions to create pointers
func stringPtr(s string) *string {
	return &s
}
