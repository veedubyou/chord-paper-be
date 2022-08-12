package publish

import (
	"github.com/rabbitmq/amqp091-go"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/lib/cerr"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

var _ Publisher = RabbitMQPublisher{}

//counterfeiter:generate . Publisher
type Publisher interface {
	Publish(msg amqp091.Publishing) error
}

func NewRabbitMQPublisher(conn *amqp091.Connection, queueName string) (RabbitMQPublisher, error) {
	channel, err := conn.Channel()
	if err != nil {
		return RabbitMQPublisher{}, cerr.Wrap(err).Error("Failed to create rabbit channel")
	}

	return RabbitMQPublisher{
		channel:   channel,
		queueName: queueName,
	}, nil
}

type RabbitMQPublisher struct {
	channel   *amqp091.Channel
	queueName string
}

func (r RabbitMQPublisher) Publish(msg amqp091.Publishing) error {
	msg.ContentType = "application/json"
	msg.DeliveryMode = amqp091.Persistent
	return r.channel.Publish("", r.queueName, true, false, msg)
}
