package service

import (
	"notebook-service/api/controller"
	"notebook-service/internal/mongo_repository"
	"notebook-service/internal/rabbitmq"
	"notebook-service/redis_repository"
	"os"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type NotebookService struct {
	rbmq      rabbitmq.RabbitMQHandler
	redisRepo redis_repository.NotebookRepository
	mongoRepo mongo_repository.NotebookRepository
	controller.UnimplementedNotebookServiceServer
}

func GenerateNotebookService(rbmq rabbitmq.RabbitMQHandler, redisRepo redis_repository.NotebookRepository, mongoRepo mongo_repository.NotebookRepository) controller.NotebookServiceServer {
	// Set the message handlers
	handlers := map[string]func([]byte){
		rabbitmq.GenerateRoutingKey(rabbitmq.PVC, rabbitmq.DELETE): HandlePVCDeleted,
	}

	go rbmq.ConsumeMessages(handlers)

	return &NotebookService{rbmq: rbmq, mongoRepo: mongoRepo, redisRepo: redisRepo}
}

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
	config := Configuration{}

	config.UrlBase = getEnvironmentVariable("URLBASE")
	config.Namespace = getEnvironmentVariable("NAMESPACE")
	config.KubeflowKustomizationPath = getEnvironmentVariable("KUBEFLOW_KUSTOMIZATION_PATH")

	return config
}
