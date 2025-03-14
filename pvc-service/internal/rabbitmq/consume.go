package rabbitmq

import (
	"log"
)

// Method to setup the queue
func (rbmq *rabbitMQHandler) setupQueue() string {
	// Declare the exchange
	err := rbmq.channel.ExchangeDeclare(
		EXCHANGE_NAME,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed declaring exchange: %v", err)
	}

	// Declare the queues
	q, err := rbmq.channel.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed declaring queue: %v", err)
	}

	// Bind the queue
	rbmq.bindQueue(q.Name, NOTEBOOK+"."+DELETE)
	rbmq.bindQueue(q.Name, NOTEBOOK+"."+CREATE)

	return q.Name
}

// method to bind the queue with routing key
func (rbmq *rabbitMQHandler) bindQueue(queue, key string) {
	err := rbmq.channel.QueueBind(
		queue,
		key,
		EXCHANGE_NAME,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed binding queue to routing key %s: %v", key, err)
	}
}

// Method to consume messages
func (rbmq *rabbitMQHandler) ConsumeMessages(handlers map[string]func([]byte)) {
	// Setup the queue
	queue := rbmq.setupQueue()

	// Consume messages
	msgs, err := rbmq.channel.Consume(
		queue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed consuming messages: %v", err)
	}

	go func() {
		for d := range msgs {
			// Get the appropriate method
			handler := handlers[d.RoutingKey]

			// Run the handler
			handler(d.Body)
		}
	}()
}
