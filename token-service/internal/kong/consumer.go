package kong

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-resty/resty/v2"
)

// Kong consumer
type Consumer struct {
	Username string   `json:"username"`
	CustomID string   `json:"custom_id"`
	Tags     []string `json:"tags,omitempty"`
}

// Method to register username as kong's consumer
var RegisterConsumer = func(username string) error {
	// Create new consumer object
	consumer := Consumer{
		Username: username,
		CustomID: username,
	}

	// encode the consumer as json
	consumerPayload, err := json.Marshal(consumer)

	if err != nil {
		return err
	}

	// Get kong admin's consumer url
	url := os.Getenv("KONG_CONSUMER_ADMIN_URI")

	// Create resty client
	client := resty.New()

	// Send the payload to kong admin
	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(consumerPayload).
		Post(url)

	if err != nil {
		log.Println(url)
		return err
	}

	// Check if registration is succesful
	status := response.StatusCode()
	if status != http.StatusCreated && status != http.StatusConflict {
		return err
	}

	// Get jwt secret from env
	secret := os.Getenv("SECRET_KEY")

	// Set the form data
	formData := map[string]string{
		"key":    username,
		"secret": secret,
	}

	// Enable jwt plugin for kong consumer
	_, err = client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(formData).
		Post(fmt.Sprintf("%s/%s/jwt", url, username))

	return err
}
