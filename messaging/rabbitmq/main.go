package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	allMessagesTopic  = "all-messages.topic"
	processedExchange = "processed.direct"
)

var (
	ch *amqp.Channel

	allMessagesQueue amqp.Queue
	plainTextQueue   amqp.Queue
	jsonQueue        amqp.Queue
)

func consumeAllMessages(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	msgs, err := ch.Consume(allMessagesQueue.Name, "all-messages-consumer", true, false, false, false, nil)
	failOnError(err, "failed to consume messages from 'all-messages.queue'")

	for {
		select {
		case <-ctx.Done():
			slog.Debug("context cancelled")
			return
		case msg := <-msgs:
			slog.Debug("received message", "body", string(msg.Body))
			if err = json.Unmarshal(msg.Body, &map[string]interface{}{}); err != nil {
				if err = ch.Publish(processedExchange, "plain", false, false, amqp.Publishing{
					ContentType: "text/plain",
					Body:        msg.Body,
				}); err != nil {
					slog.Error("failed to publish message", "error", err)
				}
				continue
			}

			if err = ch.Publish(processedExchange, "json", false, false, amqp.Publishing{
				ContentType: "text/plain",
				Body:        msg.Body,
			}); err != nil {
				slog.Error("failed to publish message", "error", err)
			}
		}
	}
}

func consumeProcessed(ctx context.Context, wg *sync.WaitGroup, queue amqp.Queue) {
	defer wg.Done()
	msgs, err := ch.Consume(queue.Name, queue.Name+"-consumer", true, false, false, false, nil)
	failOnError(err, "failed to consume messages from "+queue.Name)

	for {
		select {
		case <-ctx.Done():
			slog.Debug("context cancelled")
			return
		case msg := <-msgs:
			slog.Info("received message", "queue", queue.Name, "body", string(msg.Body))
		}
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err = conn.Channel()
	failOnError(err, "failed to open a channel")
	defer ch.Close()

	setup(ch)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go consumeAllMessages(ctx, wg)
	wg.Add(1)
	go consumeProcessed(ctx, wg, plainTextQueue)
	wg.Add(1)
	go consumeProcessed(ctx, wg, jsonQueue)

	<-sigchan
	slog.Info("termination signal received")
	cancel()

	wg.Wait()
}

func setup(ch *amqp.Channel) {
	slog.Info("setting up RabbitMQ")

	failOnError(ch.ExchangeDeclare(allMessagesTopic, "topic", true, false, false, false, nil), "failed to declare 'all-messages.topic' exchange")
	failOnError(ch.ExchangeDeclare(processedExchange, "direct", true, false, false, false, nil), "failed to declare 'processed.direct' exchange")

	var err error
	allMessagesQueue, err = ch.QueueDeclare("all-messages.queue", true, false, false, false, nil)
	failOnError(err, "failed to declare 'all-messages.queue' queue")
	failOnError(ch.QueueBind(allMessagesQueue.Name, "#", allMessagesTopic, false, nil), "failed to bind 'all-messages.queue' to 'all-messages.topic' exchange")

	plainTextQueue, err = ch.QueueDeclare("plain-text.queue", true, false, false, false, nil)
	failOnError(err, "failed to declare 'plain-text.queue' queue")
	failOnError(ch.QueueBind(plainTextQueue.Name, "plain", processedExchange, false, nil), "failed to bind 'plain-text.queue' to 'processed.direct' exchange with 'plain' routing key")

	jsonQueue, err = ch.QueueDeclare("json.queue", true, false, false, false, nil)
	failOnError(err, "failed to declare 'json.queue' queue")
	failOnError(ch.QueueBind(jsonQueue.Name, "json", processedExchange, false, nil), "failed to bind 'json.queue' to 'processed.direct' exchange with 'json' routing key")

	slog.Info("RabbitMQ setup complete")
}

func failOnError(err error, msg string) {
	if err != nil {
		panic(msg + ": " + err.Error())
	}
}
