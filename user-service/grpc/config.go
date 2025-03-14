package grpc

import (
	"fmt"
	"log"
	"net"
	"user-service/api/controller"

	"google.golang.org/grpc"
)

func SetupGRPCServer(pvcService controller.UserServiceServer) {
	// Listen the tcp port for grpc server
	port := "50051"
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed listening to tcp port %s: %v", port, err)
	}

	// Create a new gRPC server
	server := grpc.NewServer()

	// Register the service
	controller.RegisterUserServiceServer(server, pvcService)

	// Run the server
	log.Printf("gRPC server running on port %s", port)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Unable to run gRPC server on port %s", port)
	}

}
