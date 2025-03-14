package db

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var db *redis.Client

// Method to setup the redis database
func Setup(c context.Context) *redis.Client {
	url := os.Getenv("REDIS_URL")
	password := os.Getenv("REDIS_PASSWORD")

	// Connect to redis
	db = redis.NewClient(&redis.Options{
		Addr:     url,
		Password: password,
		DB:       0,
	})

	// Check connectivity
	err := db.Ping(c).Err()

	if err != nil {
		log.Fatalf("Could not connect to redis: %v", err)
	}

	return db
}
