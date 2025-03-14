package service

import (
	"context"
	"fmt"
	"log"
	"pvc-service/api/controller"
	"pvc-service/internal/rabbitmq"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// DeletePvc deletes a PVC in the specified namespace and publishes an event to RabbitMQ
func (s *PVCService) DeletePvc(ctx context.Context, req *controller.DeletePvcRequest) (*emptypb.Empty, error) {
	// Get Kubernetes configuration
	config, err := getKubeConfigFunc()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed getting kube config")
	}

	// Create a dynamic client to interact with Kubernetes resources
	client, err := CreateDynamicClient(config)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed creating dynamic client")
	}

	// Extract the PVC name from the request
	pvcName := req.Name + "-workspace"

	// Define the GroupVersionResource for PVCs
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "persistentvolumeclaims",
	}

	// Get namespace from configuration
	environmentConfig := GetConfiguration()
	namespace := environmentConfig.Namespace

	// Delete the PVC by name in the specified namespace
	err = client.Resource(gvr).Namespace(namespace).Delete(ctx, pvcName, metav1.DeleteOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error deleting PVC: %v", err)
	}

	// Publish a message to RabbitMQ
	message := fmt.Sprintf("{\"pvc_name\": \"%s\"}", pvcName)
	key := rabbitmq.GenerateRoutingKey(rabbitmq.PVC, rabbitmq.DELETE)
	log.Println(key)
	err = s.rbmq.Publish(key, message)
	if err != nil {
		log.Printf("Failed to publish PVC deletion message: %v", err)
		return nil, status.Errorf(codes.Internal, "error publishing RabbitMQ message: %v", err)
	}

	// Delete the PVC in the database

	//remove the workspace suffix
	pvcName = pvcName[:len(pvcName)-len("-workspace")]
	err = s.db.DeletePvc(pvcName)
	if err != nil {
		log.Printf("Failed to delete PVC from database: %v", err)
		log.Printf("Warning: error deleting PVC from database: %v", err)
	}

	fmt.Printf("PVC '%s' deleted successfully.\n", pvcName)
	return &emptypb.Empty{}, nil
}
