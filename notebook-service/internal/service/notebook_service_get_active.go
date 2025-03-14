package service

import (
	"context"
	"log"
	"notebook-service/api/controller"
	"notebook-service/internal/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *NotebookService) ListActiveNotebooks(ctx context.Context, request *controller.ListActiveNotebooksRequest) (*controller.ListActiveNotebooksResponse, error) {
	username := ctx.Value(auth.CtxKey).(string)

	// Check if cache exists
	cacheExists, err := s.redisRepo.CheckCacheExists(username)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var notebooks []string

	if cacheExists {
		// Get notebook from cache if cache exists
		notebooks, err = s.redisRepo.GetNotebooks(username)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		// If cache does not exists
		// Get the notebook list from mongodb
		notebooks, err = s.mongoRepo.ListNotebooks(username)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		go func() {
			// Cache the notebook list
			err = s.redisRepo.StoreNotebooks(username, notebooks)
			if err != nil {
				log.Println(err.Error())
			}
		}()
	}

	return &controller.ListActiveNotebooksResponse{
		NotebookNames: notebooks,
	}, nil
}
