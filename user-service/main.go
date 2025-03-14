package main

import (
	"log"
	"os"
	"user-service/grpc"
	"user-service/internal/service"

	"github.com/joho/godotenv"
)

func main() {
	userService := service.CreateUserService()
	grpc.SetupGRPCServer(userService)
}

func init() {
	if os.Getenv("MODE") != "PROD" {
		// Load the environment variables
		err := godotenv.Load()

		if err != nil {
			log.Fatalf("failed to load environment variables: %v", err)
		}
	}
}
