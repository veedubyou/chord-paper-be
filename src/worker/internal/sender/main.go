package main

import (
	"encoding/json"
	"github.com/rabbitmq/amqp091-go"
	"github.com/veedubyou/chord-paper-be/src/shared/config/envvar"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/job_message"
	"github.com/veedubyou/chord-paper-be/src/worker/internal/application/jobs/start"
)

func main() {
	rabbitURL := envvar.MustGet(envvar.RABBITMQ_URL)

	conn, err := amqp091.Dial(rabbitURL)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	rabbitChannel, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer rabbitChannel.Close()

	queue, err := rabbitChannel.QueueDeclare(
		"chord-paper-dev",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		panic(err)
	}

	startJobParams := start.JobParams{
		TrackIdentifier: job_message.TrackIdentifier{
			TrackListID: "ad2fca6d-8c32-4030-86c0-8b5339347253",
			TrackID:     "440a7737-bcda-4761-ae89-d85880f4bce3",
		},
	}

	jobBody, err := json.Marshal(startJobParams)

	if err != nil {
		panic(err)
	}

	job := amqp091.Publishing{Type: start.JobType, Body: jobBody}

	job.DeliveryMode = amqp091.Persistent
	job.ContentType = "application/json"

	err = rabbitChannel.Publish("", queue.Name, true, false, job)

	if err != nil {
		panic(err)
	}
}
