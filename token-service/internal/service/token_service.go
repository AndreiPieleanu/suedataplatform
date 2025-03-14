package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"token-service/api/controller"
	"token-service/internal/bcrypt"
	"token-service/internal/jwt"
	"token-service/internal/kong"
	"token-service/internal/model"
	"token-service/internal/repository"

	"github.com/redis/go-redis/v9"
)

type TokenService struct {
	controller.UnimplementedTokenServer
	tokenRepo repository.TokenRepository
}

// Method to create a token service
func CreateTokenService(tokenRepo repository.TokenRepository) controller.TokenServer {
	return &TokenService{tokenRepo: tokenRepo}
}

// Function to get or generate token from provided username
func (tokenService *TokenService) Login(context context.Context, request *controller.LoginRequest) (*controller.LoginResponse, error) {
	username := request.Username

	// Get token from database
	user, err := tokenService.tokenRepo.FindUserByUsername(username)

	if err == redis.Nil {
		// hash the password
		hashedPassword, err := bcrypt.Hash(request.Password)

		if err != nil {
			return nil, fmt.Errorf("failed hashing password: %v", err)
		}

		// Check role validity
		if request.Role == controller.Role_UNKNOWN.Enum() {
			return nil, errors.New("invalid role provided")
		}

		var wg sync.WaitGroup
		errChan := make(chan error)

		wg.Add(1)
		go func() {
			defer wg.Done()

			// Register a kong customer
			kongErr := kong.RegisterConsumer(username)
			if kongErr != nil {
				errChan <- fmt.Errorf("failed registering a kong consumer: %v", kongErr)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			// Insert the new user data to database
			user = &model.User{
				Password: hashedPassword,
				Role:     request.Role,
			}
			redisErr := tokenService.tokenRepo.CreateUser(username, user)

			if redisErr != nil {
				errChan <- fmt.Errorf("failed storing the token: %v", redisErr)
			}
		}()

		go func() {
			wg.Wait()
			close(errChan)
		}()

		for err := range errChan {
			return nil, err
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed retrieving token in database: %v", err)
	} else {
		if !bcrypt.Compare(request.Password, user.Password) {
			return nil, errors.New("invalid password provided")
		}
	}

	// Get role
	role := user.Role.String()

	if role == "UNKNOWN" {
		return nil, errors.New("invalid role")
	}

	// Generate new token if username not found in database
	token, err := jwt.GenerateToken(username, role)

	if err != nil {
		return nil, fmt.Errorf("failed generating a token: %v", err)
	}

	return &controller.LoginResponse{Token: token}, nil
}
