package mock_repository

import (
	"github.com/stretchr/testify/mock"
)

// Mocked PVC repository
type PvcRepositoryMock struct {
	mock.Mock
}

func (m *PvcRepositoryMock) CachePvcList(pvcList []string) error {
	args := m.Called(pvcList)
	return args.Error(0)
}

func (m *PvcRepositoryMock) CheckPvcExistsInCache(pvcName string) (bool, error) {
	args := m.Called(pvcName)
	return args.Bool(0), args.Error(1)
}

func (m *PvcRepositoryMock) CreatePvc(pvcName string) error {
	args := m.Called(pvcName)
	return args.Error(0)
}

func (m *PvcRepositoryMock) DeletePvc(pvcName string) error {
	args := m.Called(pvcName)
	return args.Error(0)
}
