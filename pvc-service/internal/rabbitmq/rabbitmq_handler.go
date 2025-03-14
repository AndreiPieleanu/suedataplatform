package rabbitmq

import (
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
)

type RabbitMQHandler interface {
	Publish(string, string) error
	ConsumeMessages(map[string]func([]byte))
	Close()
}

// RabbitMQHandler handles RabbitMQ connections and operations
type rabbitMQHandler struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

const EXCHANGE_NAME = "suedataplatform"

const (
	PVC      = "PVC"
	NOTEBOOK = "NOTEBOOK"

	CREATE = "CREATE"
	DELETE = "DELETE"
)

// NewRabbitMQHandler initializes and returns a RabbitMQHandler
func NewRabbitMQHandler() RabbitMQHandler {
	rabbitmqUrl := os.Getenv("RABBIT_MQ_URL")
	username := os.Getenv("RABBIT_MQ_USERNAME")
	password := os.Getenv("RABBIT_MQ_PASSWORD")

	url := fmt.Sprintf("amqp://%s:%s@%s/", username, password, rabbitmqUrl)
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Panicf("Failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Panicf("Failed to open a RabbitMQ channel: %v", err)
	}

	err = ch.ExchangeDeclare(
		EXCHANGE_NAME,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Panicf("Failed to declare an exchange: %v", err)
	}

	return &rabbitMQHandler{
		conn:    conn,
		channel: ch,
	}
}

// Publish publishes a message to the queue
func (r *rabbitMQHandler) Publish(key, message string) error {
	err := r.channel.Publish(
		EXCHANGE_NAME, // exchange
		key,           // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(message),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message to RabbitMQ: %w", err)
	}

	log.Printf("Message published: %s", message)
	return nil
}

// Close the RabbitMQ connection and channel
func (r *rabbitMQHandler) Close() {
	if err := r.channel.Close(); err != nil {
		log.Printf("Failed to close RabbitMQ channel: %v", err)
	}
	if err := r.conn.Close(); err != nil {
		log.Printf("Failed to close RabbitMQ connection: %v", err)
	}
}

// Generate routing key
func GenerateRoutingKey(msgType string, action string) string {
	return fmt.Sprintf("%s.%s", msgType, action)
}
