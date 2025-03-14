package service_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"token-service/api/controller"
	"token-service/internal/bcrypt"
	"token-service/internal/jwt"
	"token-service/internal/kong"
	"token-service/internal/model"
	"token-service/internal/service"
	"token-service/test/mock/mock_repository"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var tokenRepoMock *mock_repository.TokenRepositoryMock
var tokenService controller.TokenServer

// Setup the test for create user
func TestMain(m *testing.M) {
	tokenRepoMock = new(mock_repository.TokenRepositoryMock)
	tokenService = service.CreateTokenService(tokenRepoMock)

	code := m.Run()

	os.Exit(code)
}

// Test create token method with failed token retrieval
func TestCreateToken_FailedRetrievingTokenByUsername(t *testing.T) {
	request := &controller.LoginRequest{
		Username: "user",
		Password: "user",
		Role:     controller.Role_ADMIN.Enum(),
	}

	// Mock the token retrieval using username operation
	errorMessage := "db error"
	tokenRepoMock.On("FindUserByUsername", request.Username).Return(nil, errors.New(errorMessage)).Once()

	// Execute the method
	response, err := tokenService.Login(context.Background(), request)

	// Assertions
	assert.Nil(t, response)
	assert.Error(t, err, fmt.Sprintf("failed retrieving token in database: %s", errorMessage))
}

// method to mock bcrypt hash method
func mockHash(hashedPassword string, err error) func() {
	// Create the mock function
	newFunc := func(string) (string, error) {
		return hashedPassword, err
	}

	// Swap the methods
	oldFunc := bcrypt.Hash
	bcrypt.Hash = newFunc

	// Return the method to restore the old method
	return func() {
		bcrypt.Hash = oldFunc
	}
}

// Test create token method with failed token retrieval
func TestCreateTokenFailedHashingPassword(t *testing.T) {
	request := &controller.LoginRequest{
		Username: "user",
		Password: "user",
		Role:     controller.Role_ADMIN.Enum(),
	}

	// Mock the token retrieval using username operation
	tokenRepoMock.On("FindUserByUsername", request.Username).Return(nil, redis.Nil).Once()

	// Mock bcrypt hashing
	errorMessage := "bcrypt error"
	restoreHash := mockHash("", errors.New(errorMessage))
	defer restoreHash()

	// Execute the method
	response, err := tokenService.Login(context.Background(), request)

	// Assertions
	assert.Nil(t, response)
	assert.Error(t, err, fmt.Sprintf("failed retrieving token in database: %s", errorMessage))
}

// Test create token method with failed token retrieval
func TestCreateTokenCreateUserInvalidRole(t *testing.T) {
	request := &controller.LoginRequest{
		Username: "user",
		Password: "user",
		Role:     controller.Role_UNKNOWN.Enum(),
	}
	// Mock the token retrieval using username operation
	tokenRepoMock.On("FindUserByUsername", request.Username).Return(nil, redis.Nil).Once()

	// Mock bcrypt hashing
	errorMessage := "bcrypt error"
	restoreHash := mockHash("", errors.New(errorMessage))
	defer restoreHash()

	// Execute the method
	response, err := tokenService.Login(context.Background(), request)

	// Assertions
	assert.Nil(t, response)
	assert.Error(t, err, "invalid role provided")
}

// Method to mock kong's consumer registration
func mockRegisterConsumer(err error) func() {
	// Create the mock method
	mockedRegisterConsumer := func(string) error {
		return err
	}

	// Swap the method
	oldFunc := kong.RegisterConsumer
	kong.RegisterConsumer = mockedRegisterConsumer

	// Return the method to revert the original method
	return func() {
		kong.RegisterConsumer = oldFunc
	}
}

// Test create token with failed registering a kong consumer
func TestCreateToken_FailedConsumerRegistration(t *testing.T) {
	request := &controller.LoginRequest{
		Username: "user",
		Password: "user",
		Role:     controller.Role_ADMIN.Enum(),
	}

	// Mock the token retrieval using username operation
	tokenRepoMock.On("FindUserByUsername", request.Username).Return(nil, redis.Nil).Once()

	// Mock bcrypt hashing
	hashedPassword := string(mock.AnythingOfType("string"))
	restoreHash := mockHash(hashedPassword, nil)
	defer restoreHash()

	// Mock kong's consumer registration
	errorMessage := "kong error"
	restoreConsumerFunc := mockRegisterConsumer(errors.New(errorMessage))

	// Restore the original method
	defer restoreConsumerFunc()

	// Mock storing user data
	user := &model.User{
		Password: hashedPassword,
		Role:     request.Role,
	}
	tokenRepoMock.On("CreateUser", request.Username, user).Return(nil).Once()

	// Execute the method
	response, err := tokenService.Login(context.Background(), request)

	// Assertions
	assert.Nil(t, response)
	assert.Error(t, err, fmt.Sprintf("failed registering a kong consumer: %s", errorMessage))
}

// Test create token with failed storing user to database
func TestCreateTokenFailedStoringUser(t *testing.T) {
	request := &controller.LoginRequest{
		Username: "user",
		Password: "user",
		Role:     controller.Role_ADMIN.Enum(),
	}

	// Mock the token retrieval using username operation
	tokenRepoMock.On("FindUserByUsername", request.Username).Return(nil, redis.Nil).Once()

	// Mock bcrypt hashing
	hashedPassword := string(mock.AnythingOfType("string"))
	restoreHash := mockHash(hashedPassword, nil)
	defer restoreHash()

	// Mock kong's consumer registration
	restoreConsumerFunc := mockRegisterConsumer(nil)

	// Restore the original method
	defer restoreConsumerFunc()

	// Mock storing user data
	user := &model.User{
		Password: hashedPassword,
		Role:     request.Role,
	}
	errorMessage := "redis error"
	tokenRepoMock.On("CreateUser", request.Username, user).Return(errors.New(errorMessage))

	// Execute the method
	response, err := tokenService.Login(context.Background(), request)

	// Assertions
	assert.Nil(t, response)
	assert.Error(t, err, fmt.Sprintf("failed storing the token: %s", errorMessage))
}

// method to mock bcrypt hash method
func mockCompare(isEqual bool) func() {
	// Create the mock function
	newFunc := func(string, string) bool {
		return isEqual
	}

	// Swap the methods
	oldFunc := bcrypt.Compare
	bcrypt.Compare = newFunc

	// Return the method to restore the old method
	return func() {
		bcrypt.Compare = oldFunc
	}
}

// Test create token with failed storing user to database
func TestCreateTokenInvalidPassword(t *testing.T) {
	request := &controller.LoginRequest{
		Username: "user",
		Password: "user",
		Role:     controller.Role_ADMIN.Enum(),
	}

	// Mock the token retrieval using username operation
	user := &model.User{
		Password: "password",
		Role:     request.Role,
	}
	tokenRepoMock.On("FindUserByUsername", request.Username).Return(user, nil).Once()

	// Mock bcrypt compare method
	restoreCompare := mockCompare(false)
	defer restoreCompare()

	// Execute the method
	response, err := tokenService.Login(context.Background(), request)

	// Assertions
	assert.Nil(t, response)
	assert.Error(t, err, "invalid password provided")
}

// Test create token with invalid role
func TestCreateTokenInvalidRole(t *testing.T) {
	request := &controller.LoginRequest{
		Username: "user",
		Password: "user",
		Role:     controller.Role_UNKNOWN.Enum(),
	}

	// Mock the token retrieval using username operation
	user := &model.User{
		Password: "password",
		Role:     request.Role,
	}
	tokenRepoMock.On("FindUserByUsername", request.Username).Return(user, nil).Once()

	// Mock bcrypt compare method
	restoreCompare := mockCompare(true)
	defer restoreCompare()

	// Execute the method
	response, err := tokenService.Login(context.Background(), request)

	// Assertions
	assert.Nil(t, response)
	assert.Error(t, err, "invalid role")
}

// Method to mock jwt creation method
func mockJWTGenerateToken(token string, err error) func() {
	// Create the mock method
	mockFunc := func(string, string) (string, error) {
		return token, err
	}

	// Replace the method
	oldFunc := jwt.GenerateToken
	jwt.GenerateToken = mockFunc

	return func() {
		jwt.GenerateToken = oldFunc
	}
}

// Test create token with failed token creation
func TestCreateToken_FailedCreatingToken(t *testing.T) {
	request := &controller.LoginRequest{
		Username: "user",
		Password: "user",
		Role:     controller.Role_ADMIN.Enum(),
	}

	// Mock the token retrieval using username operation
	user := &model.User{
		Password: "password",
		Role:     request.Role,
	}
	tokenRepoMock.On("FindUserByUsername", request.Username).Return(user, nil).Once()

	// Mock bcrypt compare method
	restoreCompare := mockCompare(true)
	defer restoreCompare()

	// Mock jwt creation
	errorMessage := "jwt error"
	restoreFunc := mockJWTGenerateToken("", errors.New(errorMessage))

	// Restore the original function
	defer restoreFunc()

	// Execute the method
	response, err := tokenService.Login(context.Background(), request)

	// Assertions
	assert.Nil(t, response)
	assert.Error(t, err, "invalid role")
}

// Test succesful create token
func TestCreateToken_SuccessCreateToken(t *testing.T) {
	request := &controller.LoginRequest{
		Username: "user",
		Password: "user",
		Role:     controller.Role_DS.Enum(),
	}

	// Mock the token retrieval using username operation
	user := &model.User{
		Password: "password",
		Role:     request.Role,
	}
	tokenRepoMock.On("FindUserByUsername", request.Username).Return(user, nil).Once()

	// Mock bcrypt compare method
	restoreCompare := mockCompare(true)
	defer restoreCompare()

	// Mock jwt creation
	token := "token"
	restoreFunc := mockJWTGenerateToken(token, nil)

	// Restore the original function
	defer restoreFunc()

	expectedResponse := &controller.LoginResponse{Token: token}

	// Execute the method
	actualResponse, err := tokenService.Login(context.Background(), request)

	// Assertions
	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, actualResponse)
}
