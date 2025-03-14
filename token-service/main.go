package main

import (
	"context"
	"log"
	"os"
	"token-service/api/controller"
	"token-service/db"
	"token-service/grpc"
	"token-service/internal/repository"
	"token-service/internal/service"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	if os.Getenv("MODE") != "PROD" {
		// Load the environment variables
		err := godotenv.Load()

		if err != nil {
			log.Fatalf("failed to load environment variables: %v", err)
		}
	}

	// Setup database
	context := context.Background()
	db := db.Setup(context)

	// Create the token service
	tokenService := GenerateTokenService(db, context)

	// Setup grpc server
	grpc.SetupGRPCServer(tokenService)

}

// Method to generate the token service
func GenerateTokenService(db *redis.Client, context context.Context) controller.TokenServer {
	// Create the token repository
	tokenRepo := repository.CreateTokenRepository(db, context)

	// Create and return the token service
	return service.CreateTokenService(tokenRepo)
}
