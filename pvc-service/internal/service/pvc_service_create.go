package service

import (
	"context"
	"fmt"
	"log"
	"pvc-service/api/controller"
	"pvc-service/internal/rabbitmq" // Import the RabbitMQ handler

	"pvc-service/internal"

	"gopkg.in/yaml.v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type VolumeSpec struct {
	APIVersion string      `yaml:"apiVersion"`
	Kind       string      `yaml:"kind"`
	Metadata   VolMetadata `yaml:"metadata"`
	Spec       VolSpec     `yaml:"spec"`
}

type VolMetadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

type VolSpec struct {
	AccessModes []string     `yaml:"accessModes"`
	Resources   VolResources `yaml:"resources"`
	StorageClassName string       `yaml:"storageClassName"`
}

type VolResources struct {
	Requests VolRequests `yaml:"requests"`
}

type VolRequests struct {
	Storage string `yaml:"storage"`
	
}

type Metadata struct {
	Annotations map[string]string `yaml:"annotations"`
	Generation  int               `yaml:"generation"`
	Labels      map[string]string `yaml:"labels"`
	Name        string            `yaml:"name"`
	Namespace   string            `yaml:"namespace"`
}

func CreateVolumeBytes(NotebookName, VolumeSize string) ([]byte, error) {
	volume := VolumeSpec{
		APIVersion: "v1",
		Kind:       "PersistentVolumeClaim",
		Metadata: VolMetadata{
			Name:      NotebookName + "-workspace",
			Namespace: "kubeflow-user-example-com",
		},
		Spec: VolSpec{
			AccessModes: []string{"ReadWriteOnce"},
			Resources: VolResources{
				Requests: VolRequests{
					Storage: VolumeSize + "Gi",
				},
			},
			StorageClassName: "gp2", // Set your StorageClass
		},
	}
	yamlBytes, err := yaml.Marshal(volume)
	if err != nil {
		// Return an INTERNAL error instead of panicking
		return nil, status.Errorf(codes.Internal, "Error marshalling yaml file: %v", err)
	}
	return yamlBytes, nil
}

var dynamicNewForConfig = func(config *rest.Config) (dynamic.Interface, error) {
	return dynamic.NewForConfig(config)
}

func (s *PVCService) CreateVolume(ctx context.Context, request *controller.CreatePvcRequest) (*emptypb.Empty, error) {

	environmentConfig := GetConfiguration()

	volumeName := request.Name
	storageSize := request.Size

	// Validate the volume name against Kubernetes naming conventions
	if nameError := internal.IsValidKubernetesName(volumeName); nameError != nil {
		return nil, nameError
	}
	// Check if the PVC already exists in the cache
	exists, err := s.db.CheckPvcExistsInCache(volumeName)
	if err != nil {
		log.Printf("Warning: Error checking PVC existence in the database: %v", err)
		exists = false
	}
	if exists {
		return nil, status.Errorf(codes.AlreadyExists, "PVC %s already exists", volumeName)
	}

	size := internal.IsValidSize(storageSize)

	// Validate the size format and value
	if size == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid volume size: %s", size)
	}

	config, err := getKubeConfigFunc()
	if err != nil {
		// Return an INTERNAL error with a description
		return nil, status.Errorf(codes.Internal, "failed to get kubeconfig: %v", err)
	}

	// Create a dynamic client to interact with Kubernetes resources
	dynamicClient, err := dynamicNewForConfig(config)
	if err != nil {
		// Return an INTERNAL error with a description
		return nil, status.Errorf(codes.Internal, "failed to create dynamic client: %v", err)
	}

	namespace := environmentConfig.Namespace

	yamlFile, err := CreateVolumeBytes(volumeName, size)
	if yamlFile == nil {
		return nil, status.Errorf(codes.Internal, "Error marshalling yaml file: %v", err)
	}

	obj := &unstructured.Unstructured{}
	err = yaml.Unmarshal(yamlFile, obj)
	if err != nil {
		// Return an INTERNAL error with a description
		return nil, status.Errorf(codes.Internal, "error decoding YAML: %v", err)
	}

	// Set the group version and kind manually
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "PersistentVolumeClaim",
	})

	// Set namespace and name in metadata
	metadata := map[string]interface{}{
		"name":      volumeName + "-workspace",
		"namespace": "kubeflow-user-example-com",
	}
	obj.SetNamespace("kubeflow-user-example-com")
	obj.SetName(volumeName + "-workspace")
	obj.SetUnstructuredContent(map[string]interface{}{
		"metadata": metadata,
		"spec": map[string]interface{}{
			"accessModes": []interface{}{"ReadWriteOnce"},
			"resources": map[string]interface{}{
				"requests": map[string]interface{}{
					"storage": size + "Gi",
				},	
			},
			"storageClassName": "gp2",
		},
	})

	// Define GroupVersionResource for PVCs
	groupVersionResourcePvc := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "persistentvolumeclaims",
	}

	// Apply the object to the cluster
	_, err = dynamicClient.Resource(groupVersionResourcePvc).Namespace(namespace).Create(context.Background(), obj, v1.CreateOptions{})
	if err != nil {
		// Return an INTERNAL error with a description
		return nil, status.Errorf(codes.Internal, "error applying YAML: %v", err)
	}

	fmt.Printf("PersistentVolumeClaim %s created successfully.\n", volumeName)

	// Publish a message to RabbitMQ
	message := fmt.Sprintf("{\"pvc_name\": \"%s\"}", volumeName)
	key := rabbitmq.GenerateRoutingKey(rabbitmq.PVC, rabbitmq.CREATE)
	log.Println(key)
	err = s.rbmq.Publish(key, message)
	if err != nil {
		log.Printf("Failed to publish PVC creation message: %v", err)
		return nil, status.Errorf(codes.Internal, "Error publishing RabbitMQ message: %v", err)
	}

	// create the PVC in the database
	err = s.db.CreatePvc(volumeName)
	if err != nil {
		log.Printf("Warning: Failed to create PVC in the database: %v", err)
	}

	return &emptypb.Empty{}, nil
}
