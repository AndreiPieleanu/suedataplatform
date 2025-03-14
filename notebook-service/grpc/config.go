package grpc

import (
	"log"
	"net"
	"notebook-service/api/controller"
	"notebook-service/internal/auth"
	"os"

	"google.golang.org/grpc"
)

func SetupGRPCServer(tokenService controller.NotebookServiceServer) {
	server, lis, url := CreateGRPCServer()
	// Register the service
	controller.RegisterNotebookServiceServer(server, tokenService)

	// Run the server
	log.Printf("gRPC server running on %s", url)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Unable to run gRPC server on %s", url)
	}

}

func CreateGRPCServer() (*grpc.Server, net.Listener, string) {
	// Listen the tcp port for grpc server
	url := os.Getenv("SERVER_URL")

	lis, err := net.Listen("tcp", url)
	if err != nil {
		log.Fatalf("Failed listening to tcp %s: %v", url, err)
	}

	// Create a new gRPC server
	interceptor := grpc.UnaryInterceptor(auth.AuthInterceptor)
	server := grpc.NewServer(interceptor)

	return server, lis, url
}
