package mongo_repository

import (
	"context"
	"fmt"
	"log"
	"notebook-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type notebookRepository struct {
	coll *mongo.Collection
}

// Method to create a notebook repository
func CreateNotebookRepository(db *mongo.Database) NotebookRepository {
	coll := db.Collection("notebooks")

	// Enable unique notebook name
	idxModel := mongo.IndexModel{
		Keys:    "notebookName",
		Options: options.Index().SetUnique(true),
	}

	_, err := coll.Indexes().CreateOne(context.Background(), idxModel)
	if err != nil {
		log.Fatalf("failed creating unique index for notebook name: %v", err)
	}

	return &notebookRepository{coll: coll}
}

// Checks if user is authorized to modify the notebook
func (r *notebookRepository) AuthorizedUser(username, notebookName string) (bool, error) {
	// Create filter to find the notebook
	filter := bson.M{"notebookName": notebookName, "username": username}

	// Check if document exists
	var notebook model.NotebookEntity

	err := r.coll.FindOne(context.Background(), filter).Decode(&notebook)

	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed checking user authorization: %v", err)
	}

	return true, nil
}

// Create Notebook
func (r *notebookRepository) CreateNotebook(notebook *model.NotebookEntity) error {
	_, err := r.coll.InsertOne(context.TODO(), notebook)
	if err != nil {
		return fmt.Errorf("failed inserting the notebook: %v", err)
	}

	return nil
}

// Delete notebook from MongoDB
func (r *notebookRepository) DeleteNotebook(notebookName string) error {
	// Create the filter for notebook deletion
	filter := bson.M{"notebookName": notebookName}

	// Delete the notebook
	_, err := r.coll.DeleteOne(context.TODO(), filter)
	if err != nil {
		return fmt.Errorf("failed removing notebook %s: %v", notebookName, err)
	}

	return nil
}

// Get list of notebook from MongoDB
func (r *notebookRepository) ListNotebooks(username string) ([]string, error) {
	// Create filter for the list of retrieved notebooks
	filter := bson.M{"username": username}

	// Create projection to limit the query result to notebookName field
	projection := options.Find().SetProjection(bson.M{"notebookName": 1, "_id": 0})

	// Find matching documents
	cursor, err := r.coll.Find(context.TODO(), filter, projection)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	// Parse results into a list of notebook names
	var results []string
	for cursor.Next(context.TODO()) {
		var notebook struct {
			NotebookName string `bson:"notebookName"`
		}
		if err := cursor.Decode(&notebook); err != nil {
			return nil, err
		}
		results = append(results, notebook.NotebookName)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
