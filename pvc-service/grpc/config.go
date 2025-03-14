package grpc

import (
	"log"
	"net"
	"os"
	"pvc-service/api/controller"
	"pvc-service/internal/auth"

	"google.golang.org/grpc"
)

func SetupGRPCServer(tokenService controller.PVCServiceServer) {
	server, lis, url := CreateGRPCServer()
	// Register the service
	controller.RegisterPVCServiceServer(server, tokenService)

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
