package mock_mongo

import (
	"notebook-service/internal/model"

	"github.com/stretchr/testify/mock"
)

// MockMongo is mocking repository layer of mongodb
type MockMongo struct {
	mock.Mock
}

func (r *MockMongo) AuthorizedUser(username, notebookName string) (bool, error) {
	args := r.Called(username, notebookName)
	return args.Bool(0), args.Error(1)
}

func (r *MockMongo) CreateNotebook(notebook *model.NotebookEntity) error {
	args := r.Called(notebook)
	return args.Error(0)
}

func (r *MockMongo) DeleteNotebook(notebookName string) error {
	args := r.Called(notebookName)
	return args.Error(0)
}

func (r *MockMongo) ListNotebooks(username string) ([]string, error) {
	args := r.Called(username)

	if notebooks, ok := args.Get(0).([]string); ok {
		return notebooks, args.Error(1)
	}

	return nil, args.Error(1)
}
