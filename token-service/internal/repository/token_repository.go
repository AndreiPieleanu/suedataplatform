package repository

import "token-service/internal/model"

type TokenRepository interface {
	CreateUser(username string, user *model.User) error
	FindUserByUsername(username string) (*model.User, error)
}
