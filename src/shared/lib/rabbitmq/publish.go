package rabbitmq

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/rabbitmq/amqp091-go"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

var _ Publisher = QueuePublisher{}

//counterfeiter:generate . Publisher
type Publisher interface {
	Publish(msg amqp091.Publishing) error
}

func NewQueuePublisher(conn *amqp091.Connection, queueName string) (QueuePublisher, error) {
	channel, err := conn.Channel()
	if err != nil {
		return QueuePublisher{}, errors.Wrap(err, "Failed to create rabbit channel")
	}

	_, err = channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return QueuePublisher{}, errors.Wrap(err, "Failed to declare the queue")
	}

	return QueuePublisher{
		channel:   channel,
		queueName: queueName,
	}, nil
}

type QueuePublisher struct {
	channel   *amqp091.Channel
	queueName string
}

func (p QueuePublisher) Publish(msg amqp091.Publishing) error {
	msg.ContentType = "application/json"
	msg.DeliveryMode = amqp091.Persistent

	return p.channel.PublishWithContext(
		context.Background(),
		"",
		p.queueName,
		true,
		false,
		msg,
	)
}
