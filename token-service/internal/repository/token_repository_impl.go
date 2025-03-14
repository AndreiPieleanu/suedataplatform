package repository

import (
	"context"
	"encoding/json"
	"token-service/internal/model"

	"github.com/redis/go-redis/v9"
)

type TokenRepositoryImpl struct {
	DB      *redis.Client
	context context.Context
}

// Function to create a token repository
func CreateTokenRepository(db *redis.Client, context context.Context) TokenRepository {
	return &TokenRepositoryImpl{DB: db, context: context}
}

// Function that create an entry of token in database
func (tokenRepo *TokenRepositoryImpl) CreateUser(username string, user *model.User) error {
	// Encode the user data
	payload, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return tokenRepo.DB.Set(tokenRepo.context, username, payload, 0).Err()
}

// Function that retrieve token based on username
func (tokenRepo *TokenRepositoryImpl) FindUserByUsername(username string) (*model.User, error) {
	// Get the payload
	payload, err := tokenRepo.DB.Get(tokenRepo.context, username).Result()

	if err != nil {
		return nil, err
	}

	var user model.User

	// Decode the payload
	err = json.Unmarshal([]byte(payload), &user)

	return &user, err
}
