package testing

import (
	"encoding/json"
	"github.com/rabbitmq/amqp091-go"
	"github.com/veedubyou/chord-paper-be/src/shared/lib/rabbitmq"
	"sync"
)

func MakeRabbitMQConnection() *amqp091.Connection {
	return ExpectSuccess(amqp091.Dial(RabbitMQHost))
}

func ResetRabbitMQ(conn *amqp091.Connection) {
	channel := ExpectSuccess(conn.Channel())
	ExpectSuccess(channel.QueuePurge(RabbitMQQueueName, false))
}

func AfterSuiteRabbitMQ(conn *amqp091.Connection) {
	channel := ExpectSuccess(conn.Channel())
	ExpectSuccess(channel.QueueDelete(RabbitMQQueueName, false, false, false))
}

func MakeRabbitMQPublisher(conn *amqp091.Connection) rabbitmq.QueuePublisher {
	publisher := ExpectSuccess(rabbitmq.NewQueuePublisher(conn, RabbitMQQueueName))
	return publisher
}

type ReceivedMessage struct {
	Type    string
	Message map[string]interface{}
}

type RabbitMQConsumer struct {
	channel          *amqp091.Channel
	channelLock      sync.Mutex
	queueName        string
	receivedMessages []ReceivedMessage
	err              error
}

func NewRabbitMQConsumer(conn *amqp091.Connection) RabbitMQConsumer {
	channel := ExpectSuccess(conn.Channel())

	return RabbitMQConsumer{
		channel:          channel,
		channelLock:      sync.Mutex{},
		queueName:        RabbitMQQueueName,
		receivedMessages: nil,
		err:              nil,
	}
}

func (r *RabbitMQConsumer) AsyncStart() {
	r.channelLock.Lock()
	if r.channel == nil {
		return
	}

	messageStream := ExpectSuccess(r.channel.Consume(
		r.queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	))
	r.channelLock.Unlock()

	for message := range messageStream {
		if r.err != nil {
			continue
		}

		body := map[string]interface{}{}
		err := json.Unmarshal(message.Body, &body)
		if err != nil {
			r.err = err
			continue
		}

		newMessage := ReceivedMessage{
			Type:    message.Type,
			Message: body,
		}

		r.receivedMessages = append(r.receivedMessages, newMessage)
	}
}

func (r *RabbitMQConsumer) Stop() {
	r.channelLock.Lock()
	defer r.channelLock.Unlock()
	_ = r.channel.Close()
	r.channel = nil
}

func (r *RabbitMQConsumer) Unload() ([]ReceivedMessage, error) {
	if r.err != nil {
		return nil, r.err
	}

	receivedMessages := r.receivedMessages
	r.receivedMessages = nil
	return receivedMessages, nil
}
