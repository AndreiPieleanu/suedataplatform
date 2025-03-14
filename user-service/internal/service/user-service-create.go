package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	pb "user-service/api/controller"

	"github.com/google/uuid"
)

// User struct represents the user in memory
type User struct {
	ID          string
	Username    string
	Password    string
	DisplayName string
}

// In-memory database (map) with a mutex for concurrent access
var users = make(map[string]User)
var mu sync.Mutex

// UserServiceServer implements the UserService gRPC server
type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
}

func (s *UserServiceServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// Validate the input fields
	if req.Username == "" || req.Password == "" || req.DisplayName == "" {
		return nil, errors.New("username, password, and display name must not be empty")
	}

	// Generate a new ID for the user
	userID := uuid.New().String()

	newUser := User{
		ID:          userID,
		Username:    req.Username,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	}

	// Add user to in-memory database
	mu.Lock()
	users[userID] = newUser
	mu.Unlock()

	fmt.Printf("User created: %s\n", newUser.Username)

	return &pb.CreateUserResponse{
		User: &pb.User{
			Id:          newUser.ID,
			Username:    newUser.Username,
			DisplayName: newUser.DisplayName,
		},
	}, nil
}
