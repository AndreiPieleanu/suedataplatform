package service

import (
	"context"
	"os"
	"testing"

	"pvc-service/api/controller"
	mock_dynamic "pvc-service/mocks"        // Generated mock for Kubernetes client
	mock_rbmq "pvc-service/mocks/mock_rbmq" // Import your RabbitMQ mock package here

	"pvc-service/mocks/mock_repository"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func TestMain(m *testing.M) {
	originalGetKubeConfig := getKubeConfig
	getKubeConfigFunc = func() (*rest.Config, error) {
		return &rest.Config{}, nil
	}

	defer func() {
		getKubeConfigFunc = originalGetKubeConfig
	}()

	code := m.Run()
	os.Exit(code)
}

func TestDeletePvc_Sucess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock dynamic client and mock resource client
	mockDynamicClient := mock_dynamic.NewMockInterface(ctrl)
	mockResourceClient := mock_dynamic.NewMockNamespaceableResourceInterface(ctrl)
	mockRabbitMQ := new(mock_rbmq.RabbitMQClientMock) // Instantiate the RabbitMQ client mock

	// Mock configuration
	mockConfig := Configuration{
		Namespace: "test-namespace",
	}
	GetConfiguration = func() Configuration {
		return mockConfig
	}

	// Set up PVC name and GroupVersionResource
	pvcName := "test-pvc"
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "persistentvolumeclaims",
	}
	expectedPvcName := pvcName + "-workspace" // Reflect the naming logic

	// Define expected interactions with the dynamic client and resource client
	mockDynamicClient.EXPECT().Resource(gvr).Return(mockResourceClient).Times(1)
	mockResourceClient.EXPECT().Namespace(mockConfig.Namespace).Return(mockResourceClient).Times(1)
	mockResourceClient.EXPECT().Delete(gomock.Any(), expectedPvcName, gomock.Any()).Return(nil).Times(1)

	// Set expectations for RabbitMQ publish call with the corrected key
	expectedMessage := "{\"pvc_name\": \"" + expectedPvcName + "\"}"
	mockRabbitMQ.On("Publish", "PVC.DELETE", expectedMessage).Return(nil).Once()

	// Override CreateDynamicClient to return mock dynamic client
	oldCreateDynamicClient := CreateDynamicClient
	CreateDynamicClient = func(*rest.Config) (dynamic.Interface, error) {
		return mockDynamicClient, nil
	}
	defer func() {
		CreateDynamicClient = oldCreateDynamicClient
	}()

	// Create PVCService instance with mock RabbitMQ client using the constructor

	mockRepo := &mock_repository.PvcRepositoryMock{}
	mockRepo.On("DeletePvc", pvcName).Return(nil)

	// Setup

	pvcService := NewPVCService(mockRabbitMQ, mockRepo, context.Background())

	// Prepare request
	req := &controller.DeletePvcRequest{
		Name: pvcName,
	}

	// Call DeletePvc and assert no errors occurred
	resp, err := pvcService.DeletePvc(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, &emptypb.Empty{}, resp)

	// Assert all expectations were met
	mockRabbitMQ.AssertExpectations(t)
}

func TestDeletePvc_InvalidName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock dependencies
	mockDynamicClient := mock_dynamic.NewMockInterface(ctrl)
	mockResourceClient := mock_dynamic.NewMockNamespaceableResourceInterface(ctrl)
	mockRabbitMQ := new(mock_rbmq.RabbitMQClientMock) // Instantiate the RabbitMQ client mock

	// Mock configuration
	mockConfig := Configuration{
		Namespace: "test-namespace",
	}
	GetConfiguration = func() Configuration {
		return mockConfig
	}

	// Set up invalid PVC name and GroupVersionResource
	invalidPvcName := "invalid-volume"
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "persistentvolumeclaims",
	}

	expectedPvcName := invalidPvcName + "-workspace" // Reflect the naming logic

	mockRepo := &mock_repository.PvcRepositoryMock{}
	mockRepo.On("DeletePvc", expectedPvcName).Return(nil)

	// Define expected interactions with the dynamic client for an invalid PVC
	mockDynamicClient.EXPECT().Resource(gvr).Return(mockResourceClient).Times(1)
	mockResourceClient.EXPECT().Namespace(mockConfig.Namespace).Return(mockResourceClient).Times(1)
	mockResourceClient.EXPECT().Delete(gomock.Any(), expectedPvcName, gomock.Any()).Return(assert.AnError).Times(1)

	// Override CreateDynamicClient to return mock dynamic client
	oldCreateDynamicClient := CreateDynamicClient
	CreateDynamicClient = func(*rest.Config) (dynamic.Interface, error) {
		return mockDynamicClient, nil
	}
	defer func() {
		CreateDynamicClient = oldCreateDynamicClient
	}()

	// Initialize PVCService with the mock RabbitMQ client
	pvcService := NewPVCService(mockRabbitMQ, nil, nil)

	// Prepare the deletion request with the invalid PVC name
	reqDelete := &controller.DeletePvcRequest{
		Name: invalidPvcName,
	}

	// Execute the deletion and check that an error occurs
	_, err := pvcService.DeletePvc(context.Background(), reqDelete)

	// Assert that an error was returned
	assert.Error(t, err, "Expected error when deleting invalid PVC, but got none")
}
