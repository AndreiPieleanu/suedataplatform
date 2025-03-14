package service

import (
	"context"
	"testing"

	pb "user-service/api/controller"

	"github.com/stretchr/testify/assert"
)

// Helper function to create a test request
func createTestUserRequest(username, password, displayName string) *pb.CreateUserRequest {
	return &pb.CreateUserRequest{
		Username:    username,
		Password:    password,
		DisplayName: displayName,
	}
}

// Test for creating a user successfully (Happy flow)
func TestCreateUser_Success(t *testing.T) {
	// Initialize the service
	service := &UserServiceServer{}

	// Create a valid request
	req := createTestUserRequest("john_doe", "password123", "John Doe")

	// Call the CreateUser method
	resp, err := service.CreateUser(context.Background(), req)

	// Verify that no error occurred
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Verify that the user was created with correct values
	assert.Equal(t, "john_doe", resp.User.Username)
	assert.Equal(t, "John Doe", resp.User.DisplayName)
	assert.NotEmpty(t, resp.User.Id)
}

// Test for creating a user with an empty username (Unhappy flow)
func TestCreateUser_EmptyUsername(t *testing.T) {
	// Initialize the service
	service := &UserServiceServer{}

	// Create a request with an empty username
	req := createTestUserRequest("", "password123", "John Doe")

	// Call the CreateUser method
	resp, err := service.CreateUser(context.Background(), req)

	// Verify that the error occurred (invalid username)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Test for creating a user with an empty password (Unhappy flow)
func TestCreateUser_EmptyPassword(t *testing.T) {
	// Initialize the service
	service := &UserServiceServer{}

	// Create a request with an empty password
	req := createTestUserRequest("john_doe", "", "John Doe")

	// Call the CreateUser method
	resp, err := service.CreateUser(context.Background(), req)

	// Verify that the error occurred (invalid password)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Test for creating a user with an empty display name (Unhappy flow)
func TestCreateUser_EmptyDisplayName(t *testing.T) {
	// Initialize the service
	service := &UserServiceServer{}

	// Create a request with an empty display name
	req := createTestUserRequest("john_doe", "password123", "")

	// Call the CreateUser method
	resp, err := service.CreateUser(context.Background(), req)

	// Verify that the error occurred (invalid display name)
	assert.Error(t, err)
	assert.Nil(t, resp)
}
