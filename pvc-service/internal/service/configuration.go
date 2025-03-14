package service

import (
	"context"
	"os"
	"pvc-service/api/controller"
	"pvc-service/internal/rabbitmq"
	"pvc-service/repository"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type Configuration struct {
	UrlBase                   string
	Namespace                 string
	KubeflowKustomizationPath string
}

func getEnvironmentVariable(varname string) string {
	variable, exists := os.LookupEnv(varname)
	if !exists {
		panic("Expected environment variable '" + varname + "' to exist.")
	}
	return variable
}

var CreateDynamicClient = func(config *rest.Config) (dynamic.Interface, error) {
	return dynamic.NewForConfig(config)
}

var GetConfiguration = func() Configuration {
	return Configuration{
		UrlBase:                   getEnvironmentVariable("URLBASE"),
		Namespace:                 getEnvironmentVariable("NAMESPACE"),
		KubeflowKustomizationPath: getEnvironmentVariable("KUBEFLOW_KUSTOMIZATION_PATH"),
	}
}

// PVCService represents the service for handling PVC operations
type PVCService struct {
	rbmq rabbitmq.RabbitMQHandler
	db   repository.PvcRepository
	ctx  context.Context
	controller.UnimplementedPVCServiceServer
}

// NewPVCService initializes a new PVCService with the provided RabbitMQ handler
func NewPVCService(rbmq rabbitmq.RabbitMQHandler, repo repository.PvcRepository, ctx context.Context) *PVCService {
	return &PVCService{
		rbmq: rbmq,
		db:   repo,
		ctx:  ctx,
	}
}

// CreatePVCService sets up the PVCService and starts message consumption
// CreatePVCService sets up the PVCService and starts message consumption
func CreatePVCService(rbmq rabbitmq.RabbitMQHandler, repo repository.PvcRepository, ctx context.Context) controller.PVCServiceServer {
	handlers := map[string]func([]byte){
		rabbitmq.GenerateRoutingKey(rabbitmq.NOTEBOOK, rabbitmq.DELETE): HandleNotebookDeleted,
	}

	go rbmq.ConsumeMessages(handlers)

	// Use NewPVCService to create and return the PVCService
	return NewPVCService(rbmq, repo, ctx)
}
