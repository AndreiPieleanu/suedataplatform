package service

import (
	"context"
	"fmt"
	"log"
	"notebook-service/api/controller"
	"notebook-service/internal"
	"notebook-service/internal/auth"
	"notebook-service/internal/rabbitmq"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// DeleteNotebook deletes a notebook in the specified namespace
func (s *NotebookService) DeleteNotebook(ctx context.Context, req *controller.DeleteNotebookRequest) (*emptypb.Empty, error) {
	//client := CreateDynamicClient

	config, err := internal.GetKubeConfig()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed getting kube config")
	}

	client, err := CreateDynamicClient(config)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed creating dynamic client")
	}

	// Extract the notebook name from the request
	notebookName := req.NotebookName

	// Define the GroupVersionResource for Kubeflow Notebooks
	gvr := schema.GroupVersionResource{
		Group:    KUBEFLOW_GROUP,
		Version:  KUBEFLOW_API_VERSION,
		Resource: "notebooks",
	}

	// Define the namespace
	environmentConfig := GetConfiguration()
	namespace := environmentConfig.Namespace

	// Delete the notebook by name in the specified namespace
	err = client.Resource(gvr).Namespace(namespace).Delete(context.TODO(), notebookName, metav1.DeleteOptions{})
	if err != nil {
		return nil, err
	}

	// Publish a message to RabbitMQ
	message := fmt.Sprintf("{\"notebook_name\": \"%s\"}", notebookName)
	key := rabbitmq.GenerateRoutingKey(rabbitmq.NOTEBOOK, rabbitmq.DELETE)
	log.Println(key)
	err = s.rbmq.Publish(key, message)
	if err != nil {
		log.Printf("Failed to publish notebook deletion message: %v", err)
		return nil, status.Errorf(codes.Internal, "Error publishing RabbitMQ message: %v", err)
	}

	fmt.Printf("Notebook '%s' deleted successfully.\n", notebookName)

	// Delete notebook from database
	username := ctx.Value(auth.CtxKey).(string)
	isAuthorized, err := s.mongoRepo.AuthorizedUser(username, notebookName)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !isAuthorized {
		return nil, status.Error(codes.PermissionDenied, "user is unauthorized to perform this operation")
	}

	err = s.mongoRepo.DeleteNotebook(notebookName)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Remove notebook from cache if cache exists
	exists, err := s.redisRepo.CheckCacheExists(username)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if exists {
		err = s.redisRepo.DeleteNotebook(username, notebookName)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &emptypb.Empty{}, nil
}
