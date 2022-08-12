package dummy

import (
	"github.com/rabbitmq/amqp091-go"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/application/publish"
	"github.com/veedubyou/chord-paper-be/worker/src/internal/application/worker"
)

var _ publish.Publisher = &RabbitMQ{}
var _ worker.MessageChannel = &RabbitMQ{}
var _ amqp091.Acknowledger = RabbitMQAcknowledger{}

type RabbitMQ struct {
	AckCounter     int
	NackCounter    int
	Unavailable    bool
	MessageChannel chan amqp091.Delivery
}

type RabbitMQAcknowledger struct {
	ack  func()
	nack func()
}

func NewRabbitMQ() *RabbitMQ {
	return &RabbitMQ{
		Unavailable:    false,
		MessageChannel: make(chan amqp091.Delivery, 100),
	}
}

func (r *RabbitMQ) Publish(msg amqp091.Publishing) error {
	if r.Unavailable {
		return NetworkFailure
	}

	acknowledger := RabbitMQAcknowledger{
		ack: func() {
			r.AckCounter++
		},
		nack: func() {
			r.NackCounter++
		},
	}

	r.MessageChannel <- amqp091.Delivery{
		Acknowledger:    acknowledger,
		ContentType:     msg.ContentType,
		ContentEncoding: msg.ContentEncoding,
		DeliveryMode:    msg.DeliveryMode,
		Timestamp:       msg.Timestamp,
		Type:            msg.Type,
		Body:            msg.Body,
	}
	return nil
}

func (r *RabbitMQ) Consume(_ string, _ string, _ bool, _ bool, _ bool, _ bool, _ amqp091.Table) (<-chan amqp091.Delivery, error) {
	if r.Unavailable {
		return nil, NetworkFailure
	}

	return r.MessageChannel, nil
}

func (r *RabbitMQ) Close() error {
	return nil
}

func (r RabbitMQAcknowledger) Ack(tag uint64, multiple bool) error {
	r.ack()
	return nil
}
func (r RabbitMQAcknowledger) Nack(tag uint64, multiple bool, requeue bool) error {
	r.nack()
	return nil
}
func (r RabbitMQAcknowledger) Reject(tag uint64, requeue bool) error {
	r.nack()
	return nil
}
