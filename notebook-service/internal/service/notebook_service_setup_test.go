package service_test

import (
	"context"
	"notebook-service/api/controller"
	"notebook-service/internal"
	"notebook-service/internal/auth"
	"notebook-service/internal/service"
	"notebook-service/mocks/mock_mongo"
	"notebook-service/mocks/mock_rbmq"
	"notebook-service/mocks/mock_redis"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var notebookService controller.NotebookServiceServer
var pvc *v1.PersistentVolumeClaim

var redis *mock_redis.MockRedis
var mongo *mock_mongo.MockMongo

var username = "user"

var ctxWithValue = context.WithValue(context.Background(), auth.CtxKey, username)

func TestMain(m *testing.M) {
	restoreGetConfig := mockGetConfiguration()
	defer restoreGetConfig()

	volumeSize, err := resource.ParseQuantity("2.5G")
	if err != nil {
		os.Exit(1)
	}
	pvc = &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "notebook-test" + service.WORKSPACE_SUFFIX,
			Namespace: "kubeflow-user-example-com",
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): volumeSize,
				},
			},
		},
	}

	// Create mock rabbitmq
	rbmq := new(mock_rbmq.RabbitMQClientMock)
	// Mock Redis Client
	// redisMock := new(mock_redis.MockRedisClient)
	// redisMock.On("Ping", mock.Anything).Return(nil)

	// Create mock redis
	redis = new(mock_redis.MockRedis)

	// Create mock mongodb repo
	mongo = new(mock_mongo.MockMongo)

	notebookService = service.GenerateNotebookService(rbmq, redis, mongo)

	rbmq.On("Publish", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
	rbmq.On("ConsumeMessages", mock.Anything).Return()

	code := m.Run()

	os.Exit(code)
}

func createFakeDynamicClient() *fake.FakeDynamicClient {
	// Create a scheme and add corev1 to it
	scheme := runtime.NewScheme()
	v1.AddToScheme(scheme)

	return fake.NewSimpleDynamicClient(scheme)
}

func mockGetKubeConfig(config *rest.Config, err error) func() {
	newFunc := func() (*rest.Config, error) {
		return config, err
	}
	oldGetKubeConfig := internal.GetKubeConfig
	internal.GetKubeConfig = newFunc

	return func() {
		internal.GetKubeConfig = oldGetKubeConfig
	}
}

const NAMESPACE = "kubeflow-user-example-com-test"

func mockGetConfiguration() func() {
	newFunc := func() service.Configuration {
		return service.Configuration{
			UrlBase:                   "http://localhost:8080",
			Namespace:                 NAMESPACE,
			KubeflowKustomizationPath: "/path/to/kustomization",
		}
	}

	oldFunc := service.GetConfiguration
	service.GetConfiguration = newFunc
	return func() {
		service.GetConfiguration = oldFunc
	}
}

func mockCreateDynamicClient(err error) func() {
	var client *fake.FakeDynamicClient

	if err == nil {
		client = createFakeDynamicClient()
	} else {
		client = nil
	}

	newFunc := func(config *rest.Config) (dynamic.Interface, error) {
		return client, err
	}

	oldFunc := service.CreateDynamicClient
	service.CreateDynamicClient = newFunc
	return func() {
		service.CreateDynamicClient = oldFunc
	}
}

func mockCallOpen() func() {
	newFunc := func(string, bool) {}

	oldFunc := service.CallOpen
	service.CallOpen = newFunc
	return func() {
		service.CallOpen = oldFunc
	}
}

func mockCreatePvcResource(sentPVC *v1.PersistentVolumeClaim, err error) func() {
	newFunc := func(*kubernetes.Clientset, string, string, resource.Quantity) (*v1.PersistentVolumeClaim, error) {
		return sentPVC, err
	}

	oldFunc := service.CreatePvcResource
	service.CreatePvcResource = newFunc

	return func() {
		service.CreatePvcResource = oldFunc
	}
}
