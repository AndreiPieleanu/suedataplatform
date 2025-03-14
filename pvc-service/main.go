package main

import (
	"context"
	"log"
	"os"
	"pvc-service/db"
	"pvc-service/grpc"
	"pvc-service/internal/rabbitmq"
	"pvc-service/internal/service"
	"pvc-service/repository"

	"fmt"

	"github.com/redis/go-redis/v9"

	"pvc-service/api/controller"

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

	rabbitMQ := rabbitmq.NewRabbitMQHandler()
	defer rabbitMQ.Close()

	context := context.Background()
	db := db.Setup(context)
	pvcRepository := GeneratePvcRepository(db, context)

	pvcService := service.CreatePVCService(rabbitMQ, pvcRepository, context)

	listPvcResponse, err := pvcService.ListPVCS(context, &controller.ListPvcRequest{})
	if err != nil {
		log.Fatalf("Failed to fetch PVCs from Kubernetes: %v", err)
	}

	fmt.Printf("Fetched PVCs: %v\n", listPvcResponse)

	pvcList := listPvcResponse.PvcNames

	// Cache the PVC list in Redis
	err = pvcRepository.CachePvcList(pvcList)
	if err != nil {
		log.Fatalf("Failed to cache PVC list in Redis: %v", err)
	}
	fmt.Println("PVC list successfully cached in Redis.")

	grpc.SetupGRPCServer(pvcService)
}

func GeneratePvcRepository(db *redis.Client, context context.Context) repository.PvcRepository {

	return repository.CreatePvcRepository(db, context)
}
