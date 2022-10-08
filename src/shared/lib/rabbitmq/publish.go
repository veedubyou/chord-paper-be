package rabbitmq

import (
	"context"
	"github.com/apex/log"
	"github.com/cockroachdb/errors"
	"github.com/rabbitmq/amqp091-go"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

var _ Publisher = &QueuePublisher{}

//counterfeiter:generate . Publisher
type Publisher interface {
	Publish(msg amqp091.Publishing) error
}

func NewQueuePublisher(rabbitMQURL string, queueName string) (*QueuePublisher, error) {
	publisher := &QueuePublisher{
		rabbitMQURL: rabbitMQURL,
		queueName:   queueName,
		channel:     nil,
	}

	err := publisher.connectChannel()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to connect to RabbitMQ")
	}

	return publisher, nil
}

type QueuePublisher struct {
	rabbitMQURL string
	channel     *amqp091.Channel
	queueName   string
}

func (q *QueuePublisher) connectChannel() error {
	q.channel = nil

	conn, err := amqp091.Dial(q.rabbitMQURL)
	if err != nil {
		return errors.Wrap(err, "Failed to dial rabbitMQURL")
	}

	channel, err := conn.Channel()
	if err != nil {
		return errors.Wrap(err, "Failed to create rabbit channel")
	}

	_, err = channel.QueueDeclare(
		q.queueName,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return errors.Wrap(err, "Failed to declare the queue")
	}

	q.channel = channel
	return nil
}

func (q *QueuePublisher) publishWithoutRetry(msg amqp091.Publishing) error {
	msg.ContentType = "application/json"
	msg.DeliveryMode = amqp091.Persistent

	return q.channel.PublishWithContext(
		context.Background(),
		"",
		q.queueName,
		true,
		false,
		msg,
	)
}

func (q *QueuePublisher) Publish(msg amqp091.Publishing) error {
	err := q.publishWithoutRetry(msg)

	if err != nil {
		publishErr := errors.Wrap(err, "Failed to publish message to rabbitMQ channel")
		shouldReset := errors.Is(err, amqp091.ErrClosed)
		if !shouldReset {
			return publishErr
		}

		err = q.connectChannel()
		if err != nil {
			log.WithError(err).
				Error("Unable to reconnect to rabbitMQ channel")
			return publishErr
		}

		return q.publishWithoutRetry(msg)
	}

	return nil
}
