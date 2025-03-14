package main

import (
	"context"
	"log"
	"notebook-service/db"
	"notebook-service/grpc"
	"notebook-service/internal/mongo_repository"
	"notebook-service/internal/rabbitmq"
	"notebook-service/internal/service"
	"notebook-service/redis_repository"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if os.Getenv("MODE") != "PROD" {
		// Load the environment variables
		err := godotenv.Load()

		if err != nil {
			log.Fatalf("failed to load environment variables: %v", err)
		}
	}
	rbmq := rabbitmq.NewRabbitMQHandler()
	defer rbmq.Close()

	// Create redis connection
	ctx := context.Background()
	redisClient := db.Setup(ctx)
	if redisClient == nil {
		log.Printf("Error regarding redis: redisClient is nil!")
		return
	}
	redisRepo := redis_repository.CreateNotebookRepository(redisClient, ctx)

	// Create mongo connection
	mongoDB := db.SetupMongoDB()

	// Create mongo repository
	mongoRepo := mongo_repository.CreateNotebookRepository(mongoDB)

	defer mongoDB.Client().Disconnect(context.Background())

	notebookService := service.GenerateNotebookService(rbmq, redisRepo, mongoRepo)
	//go service.ListenForPvcDeletion(rabbitmq.RabbitMQHandler{})
	grpc.SetupGRPCServer(notebookService)

}
