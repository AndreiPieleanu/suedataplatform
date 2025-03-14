package service_test

import (
	"testing"

	"notebook-service/api/controller"
	"notebook-service/internal/service"
	mock_dynamic "notebook-service/mocks" // Correct import for the generated mocks

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func TestDeleteNotebook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDynamicClient := mock_dynamic.NewMockInterface(ctrl)
	mockResourceClient := mock_dynamic.NewMockNamespaceableResourceInterface(ctrl)

	mockConfig := service.Configuration{
		Namespace: "test-namespace",
	}
	service.GetConfiguration = func() service.Configuration {
		return mockConfig
	}

	restoreGetKubeConfig := mockGetKubeConfig(&rest.Config{}, nil)
	defer restoreGetKubeConfig()

	notebookName := "test-notebook"
	gvr := schema.GroupVersionResource{
		Group:    "kubeflow.org",
		Version:  "v1",
		Resource: "notebooks",
	}

	mockDynamicClient.EXPECT().
		Resource(gvr).
		Return(mockResourceClient).
		Times(1)

	mockResourceClient.EXPECT().
		Namespace(mockConfig.Namespace).
		Return(mockResourceClient).
		Times(1)

	mockResourceClient.EXPECT().
		Delete(gomock.Any(), notebookName, gomock.Any()).
		Return(nil).
		Times(1)

	req := &controller.DeleteNotebookRequest{
		NotebookName: notebookName,
	}

	// Mock dynamic client
	oldCreateDynamicClient := service.CreateDynamicClient
	service.CreateDynamicClient = func(*rest.Config) (dynamic.Interface, error) {
		return mockDynamicClient, nil
	}

	defer func() {
		service.CreateDynamicClient = oldCreateDynamicClient
	}()

	// Mock if user authorized
	mongo.On("AuthorizedUser", username, req.NotebookName).Return(true, nil).Once()

	// Mock deleting notebook from mongoDB
	mongo.On("DeleteNotebook", notebookName).Return(nil).Once()

	// Mock deleting notebook from existing cache
	redis.On("CheckCacheExists", username).Return(true, nil).Once()
	redis.On("DeleteNotebook", username, notebookName).Return(nil).Once()

	resp, err := notebookService.DeleteNotebook(ctxWithValue, req)

	assert.NoError(t, err)
	assert.Equal(t, &emptypb.Empty{}, resp)
}
