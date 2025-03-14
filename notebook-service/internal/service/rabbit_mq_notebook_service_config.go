package service

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Connects to RabbitMQ and returns a channel and connection
func connectRabbitMQ() (*amqp.Channel, *amqp.Connection, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	return ch, conn, nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

// Listens for messages on the 'pvc-exchange' with the 'pvc.create' routing key
func ListenForPvcCreateEvents() {
	ch, conn, err := connectRabbitMQ()
	if err != nil {
		failOnError(err, "Failed to connect to RabbitMQ server!")
	}
	defer conn.Close()
	defer ch.Close()

	// Declare the exchange if it doesn't already exist
	err = ch.ExchangeDeclare(
		"pvc-exchange", // Exchange name
		"topic",        // Exchange type
		true,           // Durable
		false,          // Auto-delete
		false,          // Internal
		false,          // No-wait
		nil,            // Arguments
	)
	if err != nil {
		failOnError(err, "Failed to create exchange!")
	}

	// Declare a queue for this service to listen on
	q, err := ch.QueueDeclare(
		"notebook-service-queue", // Queue name
		true,                     // Durable
		false,                    // Delete when unused
		false,                    // Exclusive
		false,                    // No-wait
		nil,                      // Arguments
	)
	if err != nil {
		failOnError(err, "Failed to create queue!")
	}

	// Bind the queue to the exchange with the routing key 'pvc.create'
	err = ch.QueueBind(
		q.Name,         // Queue name
		"pvc.create",   // Routing key
		"pvc-exchange", // Exchange
		false,
		nil,
	)
	if err != nil {
		failOnError(err, "Failed to bind queue!")
	}

	// Consume messages from the queue
	msgs, err := ch.Consume(
		q.Name, // Queue
		"",     // Consumer name
		true,   // Auto-ack
		false,  // Exclusive
		false,  // No-local
		false,  // No-wait
		nil,    // Args
	)
	if err != nil {
		failOnError(err, "Failed to consume event!")
	}

	// Listen for messages in a goroutine
	go func() {
		for msg := range msgs {
			log.Printf("Received PVC create event: %s", msg.Body)
			handlePvcCreateEvent(msg.Body)
		}
	}()

	// Block until stopped
	log.Printf("Listening for PVC create events on 'pvc.create' routing key...")
	select {}
}

// Handles the event message when a PVC is created
func handlePvcCreateEvent(msg []byte) {
	log.Printf("Handling PVC create event: %s", msg)
	// Process the event here (e.g., trigger notebook-related operations)
}
