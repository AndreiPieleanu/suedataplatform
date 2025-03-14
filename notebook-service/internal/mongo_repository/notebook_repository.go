package mongo_repository

import (
	"notebook-service/internal/model"
)

type NotebookRepository interface {
	AuthorizedUser(string, string) (bool, error)
	CreateNotebook(notebook *model.NotebookEntity) error
	DeleteNotebook(notebookName string) error
	ListNotebooks(string) ([]string, error)
}
