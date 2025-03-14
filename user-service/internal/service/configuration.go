package service

import (
	"os"
	"user-service/api/controller"

	"github.com/joho/godotenv"
)

type Configuration struct {
	UrlBase                   string
	Namespace                 string
	KubeflowKustomizationPath string
}

func init() {
	err := godotenv.Load()
	if err != nil {
		println("gamer")
		//TODO replace this panic
		// panic(err)
	}
}

func getEnvironmentVariable(varname string) string {
	variable, exists := os.LookupEnv(varname)
	if !exists {
		panic("Expected environment variable '" + varname + "' to exist.")
	}

	return variable
}

var GetConfiguration = func() Configuration {
	config := Configuration{}

	config.UrlBase = getEnvironmentVariable("URLBASE")
	config.Namespace = getEnvironmentVariable("NAMESPACE")
	config.KubeflowKustomizationPath = getEnvironmentVariable("KUBEFLOW_KUSTOMIZATION_PATH")

	return config
}

type UserService struct {
	controller.UnimplementedUserServiceServer
}

func CreateUserService() controller.UserServiceServer {
	return &UserService{}
}
