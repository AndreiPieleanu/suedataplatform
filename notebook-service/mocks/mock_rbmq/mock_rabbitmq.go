// Package that mocks rabbitmq connections
package mock_rbmq

import (
	"github.com/stretchr/testify/mock"
)

// Mock RabbitMQ client
type RabbitMQClientMock struct {
	mock.Mock
}

// Method to publish a message. Not doing anything during test
func (rbmq *RabbitMQClientMock) Publish(queue string, message string) error {
	args := rbmq.Called(queue, message)
	return args.Error(0)
}

// Empty method for closing connection. Not used during testing
func (rbmq *RabbitMQClientMock) Close() {
	rbmq.Called()
}

// Empty method for consume messages
func (rbmq *RabbitMQClientMock) ConsumeMessages(handlers map[string]func([]byte)) {
	rbmq.Called(handlers)
}
