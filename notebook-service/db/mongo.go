// Package that handle db connection
package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

// Method to setup mongodb connection
func SetupMongoDB() *mongo.Database {
	// Set the client option
	url := os.Getenv("DB_URL")
	username := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")

	opt := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s", username, password, url))

	// Connect to mongodb
	client, err := mongo.Connect(context.TODO(), opt)
	if err != nil {
		log.Fatalf("failed connecting to mongodb: %v", err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("failed sending a ping command to mongodb: %v", err)
	}

	// Get the database instance
	dbName := os.Getenv("DB_NAME")

	return client.Database(dbName)
}

func Close() {
	if err := client.Disconnect(context.TODO()); err != nil {
		log.Fatal(err)
	}
}
