package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Notebook struct {
	ApiVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   metav1.ObjectMeta `json:"metadata"`
	Spec       NotebookSpec      `json:"spec"`
}

type NotebookSpec struct {
	Template NotebookTemplateSpec `json:"template"`
}

type NotebookTemplateSpec struct {
	Spec v1.PodSpec `json:"spec"`
}

type NotebookEntity struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	NotebookName string             `bson:"notebookName"`
	Username     string             `bson:"username"`
}
