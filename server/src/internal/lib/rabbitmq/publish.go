package rabbitmq

import (
	"context"
	"github.com/pkg/errors"
	"github.com/rabbitmq/amqp091-go"
)

func NewPublisher(conn *amqp091.Connection, queueName string) (Publisher, error) {
	channel, err := conn.Channel()
	if err != nil {
		return Publisher{}, errors.Wrap(err, "Failed to create rabbit channel")
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
		return Publisher{}, errors.Wrap(err, "Failed to declare the queue")
	}

	return Publisher{
		channel:   channel,
		queueName: queueName,
	}, nil
}

type Publisher struct {
	channel   *amqp091.Channel
	queueName string
}

func (p Publisher) Publish(msg amqp091.Publishing) error {
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
