// Package where services are mock
package mock_service

import (
	"context"
	"token-service/api/controller"

	"github.com/stretchr/testify/mock"
)

// Mock of token service
type TokenServiceMock struct {
	controller.UnimplementedTokenServer
	mock.Mock
}

// Mock CreateToken method
func (s *TokenServiceMock) Login(ctx context.Context, req *controller.LoginRequest) (*controller.LoginResponse, error) {
	args := s.Called(ctx, req)

	// Mock the response
	res := args.Get(0).(*controller.LoginResponse)

	return res, args.Error(1)
}
