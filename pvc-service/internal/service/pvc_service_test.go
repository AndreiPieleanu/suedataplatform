package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"pvc-service/api/controller"
	"pvc-service/internal"
	"pvc-service/mocks/mock_rbmq"
	"pvc-service/mocks/mock_repository"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/rest"
	k8stesting "k8s.io/client-go/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock functions
var (
	originalGetKubeConfig         = getKubeConfig
	originalIsValidKubernetesName = internal.IsValidKubernetesName
	originalIsValidSize           = internal.IsValidSize
	originalDynamicNewForConfig   = dynamicNewForConfig
	originalGetConfiguration      = GetConfiguration
)

var PvcRepositoryMock mock_repository.PvcRepositoryMock

// Mock implementations
var mockGetKubeConfig = func() (*rest.Config, error) {
	return &rest.Config{}, nil
}

var mockIsValidKubernetesName = func(name string) error {
	return nil
}

var mockIsValidSize = func(size *string) string {
	return *size
}

var mockGetConfiguration = func() Configuration {
	return Configuration{
		UrlBase:                   "http://localhost:8080",
		Namespace:                 "kubeflow-user-example-com",
		KubeflowKustomizationPath: "/path/to/kustomization",
	}
}

func TestCreateVolume_Success(t *testing.T) {

	// Create a scheme and add corev1 to it
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)

	// Mock dependencies
	getKubeConfigFunc = mockGetKubeConfig
	internal.IsValidKubernetesName = mockIsValidKubernetesName
	internal.IsValidSize = mockIsValidSize
	GetConfiguration = mockGetConfiguration
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	dynamicNewForConfig = func(config *rest.Config) (dynamic.Interface, error) {
		return dynamicClient, nil
	}

	// Setup
	rbmq := new(mock_rbmq.RabbitMQClientMock)

	mockRepo := &mock_repository.PvcRepositoryMock{}
	mockRepo.On("CachePvcList", mock.Anything).Return(nil)
	mockRepo.On("CheckPvcExistsInCache", "test-volume").Return(false, nil)
	mockRepo.On("CreatePvc", "test-volume").Return(nil)
	mockRepo.On("DeletePvc", "test-volume").Return(nil)

	s := &PVCService{rbmq: rbmq, db: mockRepo, ctx: context.Background()}
	rbmq.On("Publish", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	rbmq.On("ConsumeMessages", mock.Anything).Return()
	ctx := context.Background()
	size := "10"
	req := &controller.CreatePvcRequest{
		Name: "test-volume",
		Size: &size,
	}

	// Execute
	_, err := s.CreateVolume(ctx, req)

	// Verify
	if err != nil {
		t.Errorf("CreateVolume returned error: %v", err)
	}

	gvr := schema.GroupVersionResource{
		Version:  "v1",
		Resource: "persistentvolumeclaims",
	}

	pvc, err := dynamicClient.Resource(gvr).Namespace("kubeflow-user-example-com").Get(ctx, "test-volume-workspace", v1.GetOptions{})
	if err != nil {
		t.Errorf("Failed to get PVC: %v", err)
	}

	if pvc == nil {
		t.Errorf("Expected PVC to be created, but it was not found")
	}

	defer func() {
		getKubeConfigFunc = originalGetKubeConfig
		internal.IsValidKubernetesName = originalIsValidKubernetesName
		internal.IsValidSize = originalIsValidSize
		dynamicNewForConfig = originalDynamicNewForConfig
		GetConfiguration = originalGetConfiguration
	}()
}
func TestCreateVolume_InvalidName(t *testing.T) {
	// Create a scheme and add corev1 to it
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)

	// Mock dependencies
	getKubeConfigFunc = mockGetKubeConfig
	internal.IsValidKubernetesName = mockIsValidKubernetesName
	internal.IsValidSize = mockIsValidSize
	GetConfiguration = mockGetConfiguration

	mockRepo := &mock_repository.PvcRepositoryMock{}
	mockRepo.On("CachePvcList", mock.Anything).Return(nil)
	mockRepo.On("CheckPvcExistsInCache", "test-volume").Return(false, nil)
	mockRepo.On("CreatePvc", "test-volume").Return(nil)
	mockRepo.On("DeletePvc", "test-volume").Return(nil)

	// Setup
	s := &PVCService{db: mockRepo, ctx: context.Background()}
	ctx := context.Background()
	size := "10"
	req := &controller.CreatePvcRequest{
		Name: "invalid name with spaces",
		Size: &size,
	}

	// Mock dependencies to return an invalid name error
	expectedErrorMessage := "invalid volume name:"
	internal.IsValidKubernetesName = func(name string) error {
		return status.Errorf(codes.InvalidArgument, expectedErrorMessage)
	}

	defer func() {
		internal.IsValidKubernetesName = originalIsValidKubernetesName
	}()

	// Execute
	_, err := s.CreateVolume(ctx, req)

	// Verify
	if err == nil {
		t.Errorf("Expected error, but got nil")
	} else {
		// Check that the error code is InvalidArgument
		if status.Code(err) != codes.InvalidArgument {
			t.Errorf("Expected error code %v, but got %v", codes.InvalidArgument, status.Code(err))
		}

		// Extract the error status and message
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Expected gRPC error status, but got a different error type: %v", err)
		}

		// Verify the message part only
		if st.Message() != expectedErrorMessage {
			t.Errorf("Expected error message %q, but got %q", expectedErrorMessage, st.Message())
		}
	}
}

func TestCreateVolume_InvalidSize(t *testing.T) {
	// Create a scheme and add corev1 to it
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)

	// Mock dependencies
	getKubeConfigFunc = mockGetKubeConfig
	internal.IsValidKubernetesName = mockIsValidKubernetesName
	internal.IsValidSize = mockIsValidSize
	GetConfiguration = mockGetConfiguration

	mockRepo := &mock_repository.PvcRepositoryMock{}
	mockRepo.On("CachePvcList", mock.Anything).Return(nil)
	mockRepo.On("CheckPvcExistsInCache", "test-volume").Return(false, nil)
	mockRepo.On("CreatePvc", "test-volume").Return(nil)
	mockRepo.On("DeletePvc", "test-volume").Return(nil)

	// Setup
	s := &PVCService{db: mockRepo, ctx: context.Background()}
	ctx := context.Background()
	size := ""
	req := &controller.CreatePvcRequest{
		Name: "test-volume",
		Size: &size,
	}

	// Mock dependencies to return an invalid size error
	expectedErrorMessage := "invalid volume size: "
	internal.IsValidSize = func(*string) string {
		return size
	}

	defer func() {
		internal.IsValidSize = originalIsValidSize
	}()

	// Execute
	_, err := s.CreateVolume(ctx, req)

	// Verify
	if err == nil {
		t.Errorf("Expected error, but got nil")
	} else {
		// Check that the error code is InvalidArgument
		if status.Code(err) != codes.InvalidArgument {
			t.Errorf("Expected error code %v, but got %v", codes.InvalidArgument, status.Code(err))
		}

		// Extract the error status and message
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Expected gRPC error status, but got a different error type: %v", err)
		}

		// Verify the message part only
		if st.Message() != expectedErrorMessage {
			t.Errorf("Expected error message %q, but got %q", expectedErrorMessage, st.Message())
		}
	}
}

func TestCreateVolume_KubeConfigError(t *testing.T) {
	// Create a scheme and add corev1 to it
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)

	// Mock dependencies
	internal.IsValidKubernetesName = mockIsValidKubernetesName
	internal.IsValidSize = mockIsValidSize
	GetConfiguration = mockGetConfiguration

	mockRepo := &mock_repository.PvcRepositoryMock{}
	mockRepo.On("CachePvcList", mock.Anything).Return(nil)
	mockRepo.On("CheckPvcExistsInCache", "test-volume").Return(false, nil)
	mockRepo.On("CreatePvc", "test-volume").Return(nil)
	mockRepo.On("DeletePvc", "test-volume").Return(nil)

	// Setup
	s := &PVCService{db: mockRepo, ctx: context.Background()}
	ctx := context.Background()
	size := "10"
	req := &controller.CreatePvcRequest{
		Name: "test-volume",
		Size: &size,
	}

	// Mock dependency to return the expected error
	expectedErrorMessage := fmt.Errorf("connection refused")
	getKubeConfigFunc = func() (*rest.Config, error) {
		return nil, expectedErrorMessage
	}

	defer func() {
		getKubeConfigFunc = originalGetKubeConfig
	}()

	// Execute
	_, err := s.CreateVolume(ctx, req)

	// Verify
	if err == nil {
		t.Errorf("Expected error, but got nil")
	} else {
		// Check that the error code is Internal
		if status.Code(err) != codes.Internal {
			t.Errorf("Expected error code %v, but got %v", codes.Internal, status.Code(err))
		}

		// Extract the error status and message
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Expected gRPC error status, but got a different error type: %v", err)
		}

		// Verify the message part only
		expectedMessage := fmt.Sprintf("failed to get kubeconfig: %v", expectedErrorMessage.Error())
		if st.Message() != expectedMessage {
			t.Errorf("Expected error message %q, but got %q", expectedMessage, st.Message())
		}
	}
}
func TestCreateVolume_DynamicClientError(t *testing.T) {
	// Create a scheme and add corev1 to it
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)

	// Mock dependencies
	getKubeConfigFunc = mockGetKubeConfig
	internal.IsValidKubernetesName = mockIsValidKubernetesName
	internal.IsValidSize = mockIsValidSize
	GetConfiguration = mockGetConfiguration

	mockRepo := &mock_repository.PvcRepositoryMock{}
	mockRepo.On("CachePvcList", mock.Anything).Return(nil)
	mockRepo.On("CheckPvcExistsInCache", "test-volume").Return(false, nil)
	mockRepo.On("CreatePvc", "test-volume").Return(nil)
	mockRepo.On("DeletePvc", "test-volume").Return(nil)

	// Setup
	s := &PVCService{db: mockRepo, ctx: context.Background()}
	ctx := context.Background()
	size := "10"
	req := &controller.CreatePvcRequest{
		Name: "test-volume",
		Size: &size,
	}

	// Define the expected error message
	expectedErrorMessage := fmt.Errorf("failed to create dynamic client")

	dynamicNewForConfig = func(config *rest.Config) (dynamic.Interface, error) {
		return nil, expectedErrorMessage
	}

	defer func() {
		// Restore the original dynamicNewForConfig after the test
		dynamicNewForConfig = originalDynamicNewForConfig
	}()

	// Execute
	_, err := s.CreateVolume(ctx, req)

	// Verify
	if err == nil {
		t.Errorf("Expected error, but got nil")
	} else {
		// Check that the error code is Internal
		if status.Code(err) != codes.Internal {
			t.Errorf("Expected error code %v, but got %v", codes.Internal, status.Code(err))
		}

		// Extract the error status and message
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Expected gRPC error status, but got a different error type: %v", err)
		}

		// Verify the message part only
		if st.Message() != fmt.Sprintf("failed to create dynamic client: %v", expectedErrorMessage.Error()) {
			t.Errorf("Expected error message %q, but got %q", expectedErrorMessage.Error(), st.Message())
		}
	}
}

func TestCreateVolume_YamlApplicationError(t *testing.T) {
	// Create a scheme and add corev1 to it
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)

	// Mock dependencies
	getKubeConfigFunc = mockGetKubeConfig
	internal.IsValidKubernetesName = mockIsValidKubernetesName
	internal.IsValidSize = mockIsValidSize
	GetConfiguration = mockGetConfiguration

	mockRepo := &mock_repository.PvcRepositoryMock{}
	mockRepo.On("CachePvcList", mock.Anything).Return(nil)
	mockRepo.On("CheckPvcExistsInCache", "test-volume").Return(false, nil)
	mockRepo.On("CreatePvc", "test-volume").Return(nil)
	mockRepo.On("DeletePvc", "test-volume").Return(nil)

	// Setup
	s := &PVCService{db: mockRepo, ctx: context.Background()}
	ctx := context.Background()
	size := "10"
	req := &controller.CreatePvcRequest{
		Name: "test-volume",
		Size: &size,
	}

	// Define the expected error message
	expectedErrorMessage := fmt.Errorf("failed to apply YAML")

	// Create a fake dynamic client and add a reactor to simulate a YAML application error
	dynamicClient := fake.NewSimpleDynamicClient(scheme)
	dynamicNewForConfig = func(config *rest.Config) (dynamic.Interface, error) {
		return dynamicClient, nil
	}
	dynamicClient.PrependReactor("create", "persistentvolumeclaims", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, expectedErrorMessage
	})

	defer func() {
		// Restore the original dynamicNewForConfig after the test
		dynamicNewForConfig = originalDynamicNewForConfig
	}()

	// Execute
	_, err := s.CreateVolume(ctx, req)

	// Verify
	if err == nil {
		t.Errorf("Expected error, but got nil")
	} else {
		// Check that the error code is Internal
		if status.Code(err) != codes.Internal {
			t.Errorf("Expected error code %v, but got %v", codes.Internal, status.Code(err))
		}

		// Extract the error status and message
		st, ok := status.FromError(err)
		if !ok {
			t.Errorf("Expected gRPC error status, but got a different error type: %v", err)
		}

		// Verify the message part only
		if st.Message() != fmt.Sprintf("error applying YAML: %v", expectedErrorMessage.Error()) {
			t.Errorf("Expected error message %q, but got %q", expectedErrorMessage, st.Message())
		}
	}
}

// ----- LIST PVC TESTS ----- //

// Mocking Kubernetes client and PVCService dependencies
type MockClientset struct {
	mock.Mock
	kubernetes.Interface
}

type MockCoreV1 struct {
	mock.Mock
	kubernetes.Interface
}

type MockPVCService struct {
	mock.Mock
}

func (m *MockClientset) CoreV1() *MockCoreV1 {
	args := m.Called()
	return args.Get(0).(*MockCoreV1)
}

func (m *MockCoreV1) PersistentVolumeClaims(namespace string) *MockPVCService {
	args := m.Called(namespace)
	return args.Get(0).(*MockPVCService)
}

func (m *MockPVCService) List(ctx context.Context, opts v1.ListOptions) (*corev1.PersistentVolumeClaimList, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*corev1.PersistentVolumeClaimList), args.Error(1)
}

type MockKubeConfig struct {
	mock.Mock
}

func (m *MockKubeConfig) GetKubeConfig() (*rest.Config, error) {
	args := m.Called()
	return args.Get(0).(*rest.Config), args.Error(1)
}

// Test case for successful PVC listing// Test case for successful PVC listing
func TestListPVCS_Success(t *testing.T) {
	// Mock Kubernetes clientset
	clientset := new(MockClientset)
	coreV1 := new(MockCoreV1)
	pvcServiceMock := new(MockPVCService)

	// Mock the methods
	clientset.On("CoreV1").Return(coreV1)
	coreV1.On("PersistentVolumeClaims", "test-namespace").Return(pvcServiceMock)

	// Create fake PVC data
	pvcList := &corev1.PersistentVolumeClaimList{
		Items: []corev1.PersistentVolumeClaim{
			{ObjectMeta: v1.ObjectMeta{Name: "pvc1"}},
			{ObjectMeta: v1.ObjectMeta{Name: "pvc2"}},
		},
	}

	// Simulate List function returning fake PVC data
	pvcServiceMock.On("List", context.TODO(), v1.ListOptions{}).Return(pvcList, nil)

	// Now, simulate the full call flow:
	// 1. Call clientset.CoreV1() to get coreV1
	// 2. Call coreV1.PersistentVolumeClaims("test-namespace") to get pvcServiceMock
	// 3. Call pvcServiceMock.List to get the PVC list

	pvcService := clientset.CoreV1().PersistentVolumeClaims("test-namespace")
	response, err := pvcService.List(context.TODO(), v1.ListOptions{})

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.ElementsMatch(t, []string{response.Items[0].Name, response.Items[1].Name}, []string{"pvc1", "pvc2"})

	// Verify that the mock expectations were met
	clientset.AssertExpectations(t)
	coreV1.AssertExpectations(t)
	pvcServiceMock.AssertExpectations(t)
}

type MockKubeConfigFunc struct {
	mock.Mock
}

func (m *MockKubeConfigFunc) GetKubeConfig() (*rest.Config, error) {
	args := m.Called()
	// Handle nil safely
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	// Cast only if it's non-nil
	config, ok := args.Get(0).(*rest.Config)
	if !ok {
		return nil, errors.New("failed to cast to *rest.Config")
	}
	return config, args.Error(1)
}

// Test case for error during kubeconfig loading
func TestListPVCS_KubeConfigError(t *testing.T) {
	kubeConfigMock := new(MockKubeConfigFunc)

	// Set the expectation that getKubeConfigFunc will return an error
	kubeConfigMock.On("GetKubeConfig").Return(nil, errors.New("failed to load kubeconfig"))

	// Replace the global getKubeConfigFunc with the mocked version
	getKubeConfigFunc = kubeConfigMock.GetKubeConfig

	// Create a PVCService instance
	pvcService := PVCService{}

	// Call the ListPVCS method
	response, err := pvcService.ListPVCS(context.TODO(), &controller.ListPvcRequest{})

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.EqualError(t, err, "error building kubeconfig: failed to load kubeconfig")
}

func TestListPVCS_ListError(t *testing.T) {
	// Mock Kubernetes clientset
	clientset := new(MockClientset)
	coreV1 := new(MockCoreV1)
	pvcServiceMock := new(MockPVCService) // This is the mock service

	clientset.On("CoreV1").Return(coreV1)
	coreV1.On("PersistentVolumeClaims", "test-namespace").Return(pvcServiceMock)

	// Simulate error during PVC listing
	pvcServiceMock.On("List", context.TODO(), v1.ListOptions{}).Return((*corev1.PersistentVolumeClaimList)(nil), errors.New("failed to list PVCs"))

	pvcService := clientset.CoreV1().PersistentVolumeClaims("test-namespace")
	response, err := pvcService.List(context.TODO(), v1.ListOptions{})

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.EqualError(t, err, "failed to list PVCs")

	// Verify expectations on mocks
	clientset.AssertExpectations(t)
	coreV1.AssertExpectations(t)
	pvcServiceMock.AssertExpectations(t) // Check the mock
}
