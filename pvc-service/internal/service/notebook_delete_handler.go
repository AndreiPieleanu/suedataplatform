package service

import (
	"encoding/json"
	"fmt"
	"log"
)

// PVCDeletedEvent represents the structure of the deletion event
type NotebookDeletedEvent struct {
	NotebookName string `json:"notebook_name"`
}

// handleNotebookDeleted processes the event when a notebook is deleted
func HandleNotebookDeleted(message []byte) {
	var event NotebookDeletedEvent
	err := json.Unmarshal(message, &event)
	if err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		return
	}

	// Process the event (e.g., log it or update internal state)
	fmt.Printf("PVC service received notification: Notebook '%s' has been deleted.\n", event.NotebookName)
	// Implement additional logic here, like updating the database or cleaning up resources.
}
