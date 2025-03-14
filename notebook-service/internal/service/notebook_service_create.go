package service

import (
	"context"
	"fmt"
	"notebook-service/api/controller"
	"notebook-service/internal"
	"notebook-service/internal/auth"
	"notebook-service/internal/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sYaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

// const KUBEFLOW_GROUP = "kubeflow.org"
// const KUBEFLOW_API_VERSION = "v1"
const KUBEFLOW_NOTEBOOK_KIND = "Notebook"
const KUBEFLOW_NOTEBOOKS_RESOURCE = "notebooks"
const KUBEFLOW_HOME_DIRECTORY = "/home/jovyan"

const KUBEFLOW_JUPYTER_IMAGE = "kubeflownotebookswg/jupyter-scipy:v1.8.0-rc.0"
const KUBEFLOW_VSCODE_IMAGE = "kubeflownotebookswg/codeserver-python:v1.8.0"
const KUBEFLOW_RSTUDIO_IMAGE = "kubeflownotebookswg/rstudio-tidyverse:v1.8.0"

const WORKSPACE_SUFFIX = "-workspace"

const SERVER_TYPE_ANNOTATION = "notebooks.kubeflow.org/server-type"
const HEADERS_ANNOTATION = "notebooks.kubeflow.org/http-headers-request-set"
const URI_REWRITE_ANNOTATION = "notebooks.kubeflow.org/http-rewrite-uri"

const JUPYTER_SERVER_TYPE = "jupyter"
const VSCODE_SERVER_TYPE = "group-one"
const RSTUDIO_SERVER_TYPE = "group-two"

func setStringValue(value *string, defaultValue string) string {
	if value == nil {
		return defaultValue
	}
	return *value
}

func setBoolValue(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}

func (s *NotebookService) CreateNotebook(ctx context.Context, req *controller.CreateNotebookRequest) (*emptypb.Empty, error) {
	cpuLimitResource, err := resource.ParseQuantity(setStringValue(req.MaxCpu, "2"))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid max cpu")
	}

	cpuRequestResource, err := resource.ParseQuantity(setStringValue(req.MinCpu, "1"))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid min cpu")
	}

	memoryLimitResource, err := resource.ParseQuantity(setStringValue(req.MaxMemory, "2G"))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid max memory")

	}

	memoryRequestResource, err := resource.ParseQuantity(setStringValue(req.MinMemory, "1G"))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid min memory")

	}

	parsedVolumeSize, err := resource.ParseQuantity(setStringValue(req.Volume, "2.5G"))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid volume size")

	}

	environmentConfig := GetConfiguration()
	namespace := environmentConfig.Namespace

	config, err := internal.GetKubeConfig()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed getting kube config")
	}

	dynamicClient, err := CreateDynamicClient(config)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed creating dynamic client")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed creating new client set")
	}

	pvc := setStringValue(req.Pvc, "")

	pvcArg := pvc
	if pvcArg == "" {
		_, err = CreatePvcResource(clientset, namespace, req.Name, parsedVolumeSize)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed creating pvc")

		}
		pvcArg = req.Name + WORKSPACE_SUFFIX
	}

	_, err = createNotebookResource(dynamicClient, namespace, req.Name, req.Type, cpuLimitResource, cpuRequestResource, memoryLimitResource, memoryRequestResource, pvcArg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed creating new notebook")

	}

	fmt.Printf("Notebook '%s' created successfully. Please wait a few seconds for the notebook to start.\n", req.Name)

	// Store in database
	notebookEntity := &model.NotebookEntity{
		Username:     ctx.Value(auth.CtxKey).(string),
		NotebookName: req.Name,
	}

	err = s.mongoRepo.CreateNotebook(notebookEntity)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Update cache if exists
	exists, err := s.redisRepo.CheckCacheExists(notebookEntity.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if exists {
		err = s.redisRepo.AddNotebook(notebookEntity.Username, notebookEntity.NotebookName)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	open := setBoolValue(req.Open)
	CallOpen(req.Name, open)

	return nil, nil
}

var CreatePvcResource = func(clientset *kubernetes.Clientset, namespace string, notebookName string, volumeSize resource.Quantity) (*v1.PersistentVolumeClaim, error) {
	pvc := createNotebookPvcDefinition(namespace, notebookName, volumeSize)
	client := clientset.CoreV1().PersistentVolumeClaims(namespace)
	return client.Create(context.TODO(), &pvc, metav1.CreateOptions{})
}

func createNotebookResource(
	dynamicClient dynamic.Interface,
	namespace string,
	notebookName string,
	notebookType *controller.NotebookType,
	cpuLimitResource resource.Quantity,
	cpuRequestResource resource.Quantity,
	memoryLimitResource resource.Quantity,
	memoryRequestResource resource.Quantity,
	pvcName string,
) (*unstructured.Unstructured, error) {
	gvr := schema.GroupVersionResource{
		Group:    KUBEFLOW_GROUP,
		Version:  KUBEFLOW_API_VERSION,
		Resource: KUBEFLOW_NOTEBOOKS_RESOURCE,
	}

	notebook := createNotebookDefinition(namespace, notebookName, notebookType, cpuLimitResource, cpuRequestResource, memoryLimitResource, memoryRequestResource, pvcName)

	yaml, err := yaml.Marshal(notebook)
	if err != nil {
		return nil, err
	}

	obj := &unstructured.Unstructured{}
	decoder := k8sYaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, _, err = decoder.Decode(yaml, nil, obj)
	if err != nil {
		return nil, err
	}

	client := dynamicClient.Resource(gvr).Namespace(namespace)
	createdNotebook, err := client.Create(context.Background(), obj, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return createdNotebook, nil
}

func createNotebookPvcDefinition(namespace string, notebookName string, volumeSize resource.Quantity) v1.PersistentVolumeClaim {
	return v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      notebookName + WORKSPACE_SUFFIX,
			Namespace: namespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): volumeSize,
				},
			},
			StorageClassName: pointerToString("gp2"), // Add this
		},
	}
}

// Helper function to return a pointer to a string
func pointerToString(s string) *string {
	return &s
}

func createNotebookDefinition(
	namespace string,
	notebookName string,
	notebookType *controller.NotebookType,
	cpuLimitResource resource.Quantity,
	cpuRequestResource resource.Quantity,
	memoryLimitResource resource.Quantity,
	memoryRequestResource resource.Quantity,
	pvcName string,
) *model.Notebook {
	image, annotations := getNotebookAnnotationsAndImage(namespace, notebookName, notebookType)

	return &model.Notebook{
		ApiVersion: KUBEFLOW_GROUP + "/" + KUBEFLOW_API_VERSION,
		Kind:       KUBEFLOW_NOTEBOOK_KIND,
		Metadata: metav1.ObjectMeta{
			Name:        notebookName,
			Namespace:   namespace,
			Annotations: annotations,
		},
		Spec: model.NotebookSpec{
			Template: model.NotebookTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image:           image,
							ImagePullPolicy: v1.PullIfNotPresent,
							Name:            notebookName,
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    cpuLimitResource,
									v1.ResourceMemory: memoryLimitResource,
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    cpuRequestResource,
									v1.ResourceMemory: memoryRequestResource,
								},
							},
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      pvcName,
									MountPath: KUBEFLOW_HOME_DIRECTORY,
								},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: pvcName,
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
								},
							},
						},
					},
				},
			},
		},
	}
}

func getNotebookAnnotationsAndImage(
	namespace string,
	notebookName string,
	notebookType *controller.NotebookType,
) (string, map[string]string) {
	var image string
	annotations := map[string]string{}

	if notebookType == nil {
		// Handle the nil case, possibly by defaulting to Jupyter or returning an error
		notebookValue := controller.NotebookType_JUPITER
		notebookType = &notebookValue
	}

	switch *notebookType {
	case controller.NotebookType_JUPITER:
		image = KUBEFLOW_JUPYTER_IMAGE
		annotations[SERVER_TYPE_ANNOTATION] = JUPYTER_SERVER_TYPE
	case controller.NotebookType_VSCODE:
		image = KUBEFLOW_VSCODE_IMAGE
		annotations[SERVER_TYPE_ANNOTATION] = VSCODE_SERVER_TYPE
		annotations[URI_REWRITE_ANNOTATION] = "/"
	case controller.NotebookType_RSTUDIO:
		image = KUBEFLOW_RSTUDIO_IMAGE
		annotations[SERVER_TYPE_ANNOTATION] = RSTUDIO_SERVER_TYPE
		annotations[URI_REWRITE_ANNOTATION] = "/"
		annotations[HEADERS_ANNOTATION] = fmt.Sprintf(
			"{\"X-RStudio-Root-Path\":\"/notebook/%s/%s/\"}", namespace, notebookName)
	default:
		// Handle unknown types
		return "", nil
	}

	return image, annotations
}
