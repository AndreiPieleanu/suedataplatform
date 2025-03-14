package model

import "token-service/api/controller"

type User struct {
	Password string           `json:"password"`
	Role     *controller.Role `json:"role"`
}
