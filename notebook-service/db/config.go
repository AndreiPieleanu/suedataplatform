package db

import (
	"context"
	"fmt"
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

	// Check connectivity
	err := db.Ping(c).Err()

	if err != nil {
		fmt.Printf("Could not connect to redis: " + err.Error())
		return nil
	}

	return db
}
