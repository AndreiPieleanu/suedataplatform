package service

import (
	"encoding/json"
	"fmt"
	"log"
)

// PVCDeletedEvent represents the structure of the deletion event
type PVCDeletedEvent struct {
	PvcName string `json:"pvc_name"`
}

// handlePVCDeleted processes the event when a PVC is deleted
func HandlePVCDeleted(message []byte) {
	var event PVCDeletedEvent
	err := json.Unmarshal(message, &event)
	if err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		return
	}

	// Process the event (e.g., log it or update internal state)
	fmt.Printf("Notebook service received notification: PVC '%s' has been deleted.\n", event.PvcName)
	// Implement additional logic here, like updating the database or cleaning up resources.
}
