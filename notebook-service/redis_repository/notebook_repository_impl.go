package redis_repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type NotebookRepositoryImpl struct {
	DB      *redis.Client
	context context.Context
}

func CreateNotebookRepository(db *redis.Client, context context.Context) NotebookRepository {
	return &NotebookRepositoryImpl{DB: db, context: context}
}

// Helper to generate the key
func (r *NotebookRepositoryImpl) generateKey(username string) string {
	return "notebook:" + username
}

// CheckCacheExists checks if user's cache exists in redis
func (r *NotebookRepositoryImpl) CheckCacheExists(username string) (bool, error) {
	// Get the data key
	key := r.generateKey(username)

	exists, err := r.DB.Exists(r.context, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed checking cache existence: %v", err)
	}

	return exists > 0, nil
}

// StoreNotebooks cache the get active notebooks values to redis
func (r *NotebookRepositoryImpl) StoreNotebooks(username string, notebooks []string) error {
	// Get the data key
	key := r.generateKey(username)

	// Store the notebooks
	pipe := r.DB.Pipeline()
	pipe.SAdd(r.context, key, notebooks)
	pipe.Expire(r.context, key, 2*time.Hour)

	if _, err := pipe.Exec(r.context); err != nil {
		return fmt.Errorf("failed storing notebooks: %v", err)
	}

	return nil
}

// AddNotebook add a notebook value to existing cache
func (r *NotebookRepositoryImpl) AddNotebook(username string, notebook string) error {
	// Get the data key
	key := r.generateKey(username)

	// Store the notebooks
	if err := r.DB.SAdd(r.context, key, notebook).Err(); err != nil {
		return fmt.Errorf("failed adding notebook: %v", err)
	}

	return nil
}

// GetNotebooks retrieved cached notebook list
func (r *NotebookRepositoryImpl) GetNotebooks(username string) ([]string, error) {
	// Get cache key
	key := r.generateKey(username)

	// Get the list of notebook
	notebooks, err := r.DB.SMembers(r.context, key).Result()
	if err != nil {
		return []string{}, fmt.Errorf("failed getting the list of notebook: %v", err)
	}

	return notebooks, nil
}

// DeleteNotebook removes a notebook from Redis cache
func (r *NotebookRepositoryImpl) DeleteNotebook(username, notebook string) error {
	// Get cache key
	key := r.generateKey(username)

	// Delete the notebook from existing cache
	if err := r.DB.SRem(r.context, key, notebook).Err(); err != nil {
		return fmt.Errorf("failed deleting notebook %s: %v", notebook, err)
	}

	return nil
}
