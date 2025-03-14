package service

import (
	"context"
	"fmt"
	"pvc-service/api/controller"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (s *PVCService) ListPVCS(ctx context.Context, request *controller.ListPvcRequest) (*controller.ListPvcResponse, error) {

	// Load kubeconfig file
	config, err := getKubeConfigFunc()
	if err != nil {
		return nil, fmt.Errorf("error building kubeconfig: %v", err)
	}

	// Create Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating clientset: %v", err)
	}

	environmentConfig := GetConfiguration()
	// Specify the namespace where PVCs exist
	namespace := environmentConfig.Namespace

	// List PVCs in the namespace
	pvcs, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing PVCs: %v", err)
	}

	// Collect PVC names in a slice
	var pvcNames []string
	for _, pvc := range pvcs.Items {
		pvcNames = append(pvcNames, pvc.Name)
	}

	// Return the list of PVC names as a gRPC response
	response := &controller.ListPvcResponse{
		PvcNames: pvcNames,
	}

	return response, nil
}
