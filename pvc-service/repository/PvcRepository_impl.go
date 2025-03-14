package repository

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type PvcRepositoryImpl struct {
	DB      *redis.Client
	context context.Context
}

// Function to create a PVC repository
func CreatePvcRepository(db *redis.Client, context context.Context) PvcRepository {
	return &PvcRepositoryImpl{DB: db, context: context}
}

// Function to cache the list of PVCs in the database
func (pvcRepo *PvcRepositoryImpl) CachePvcList(pvcList []string) error {
	for _, pvcName := range pvcList {
		// Store each PVC as an individual key
		err := pvcRepo.DB.Set(pvcRepo.context, fmt.Sprintf("pvc:%s", pvcName), "true", 0).Err()
		if err != nil {
			return fmt.Errorf("failed to cache PVC %s: %w", pvcName, err)
		}
	}
	return nil
}

// Function to check if a PVC exists in the cache
func (pvcRepo *PvcRepositoryImpl) CheckPvcExistsInCache(pvcName string) (bool, error) {
	// Check if the key exists in Redis
	result, err := pvcRepo.DB.Exists(pvcRepo.context, fmt.Sprintf("pvc:%s", pvcName)).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check PVC existence: %w", err)
	}
	return result > 0, nil
}

// Function to create a PVC in the database
func (pvcRepo *PvcRepositoryImpl) CreatePvc(pvcName string) error {
	// Add a new PVC to the Redis cache
	err := pvcRepo.DB.Set(pvcRepo.context, fmt.Sprintf("pvc:%s", pvcName), "true", 0).Err()
	if err != nil {
		return fmt.Errorf("failed to create PVC %s: %w", pvcName, err)
	}
	return nil
}

// Function to delete a PVC in the database
func (pvcRepo *PvcRepositoryImpl) DeletePvc(pvcName string) error {
	// Remove the PVC from the Redis cache
	err := pvcRepo.DB.Del(pvcRepo.context, fmt.Sprintf("pvc:%s", pvcName)).Err()
	if err != nil {
		return fmt.Errorf("failed to delete PVC %s: %w", pvcName, err)
	}
	return nil
}
