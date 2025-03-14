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
	// Connect to redis
	db = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	log.Printf("address: %s", os.Getenv("REDIS_URL"))
	log.Printf("password: %s", os.Getenv("REDIS_PASSWORD"))

	// Check connectivity
	err := db.Ping(c).Err()

	if err != nil {
		log.Fatalf("Could not connect to redis: %v", err)
	}

	return db
}
