package mock_repository

import (
	"token-service/internal/model"

	"github.com/stretchr/testify/mock"
)

// Mocked token repository
type TokenRepositoryMock struct {
	mock.Mock
}

// Mock token creation in db method
func (tokenRepoMock *TokenRepositoryMock) CreateUser(username string, user *model.User) error {
	args := tokenRepoMock.Called(username, user)

	return args.Error(0)
}

// Mock the find token by username method
func (tokenRepoMock *TokenRepositoryMock) FindUserByUsername(username string) (*model.User, error) {
	args := tokenRepoMock.Called(username)

	// Mock the returned user model
	if user, ok := args.Get(0).(*model.User); ok {
		return user, args.Error(1)
	}

	return nil, args.Error(1)
}
