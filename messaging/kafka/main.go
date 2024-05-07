package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func process(ctx context.Context, producer *kafka.Writer, toProcess <-chan *kafka.Message, processed chan<- *kafka.Message) {
	defer close(processed)

	for m := range toProcess {
		target := kafka.Message{
			Key:     m.Key,
			Value:   m.Value,
			Headers: m.Headers,
			Time:    time.Now(),
		}
		var data map[string]interface{}
		if err := json.Unmarshal(m.Value, &data); err != nil {
			target.Topic = "plain_text"
		} else {
			target.Topic = "json"
		}

		if err := producer.WriteMessages(ctx, target); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
				slog.Debug("context cancelled")
				break
			}

			slog.Error("failed to write message", "error", err)
		}

		processed <- m
	}
}

func run(ctx context.Context) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "all_messages",
		GroupID:  randomString(5), // random group ID for easier testing
		MinBytes: 10e3,            // 10KB
		MaxBytes: 10e6,            // 10MB
		MaxWait:  1 * time.Second,
	})
	defer r.Close()
	w := &kafka.Writer{
		Addr: kafka.TCP("localhost:9092"),
	}
	defer w.Close()

	toProcess := make(chan *kafka.Message)
	processed := make(chan *kafka.Message)

	go process(ctx, w, toProcess, processed)

	go func() {
		defer close(toProcess)
		for {
			m, err := r.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
					slog.Debug("context cancelled")
					break
				}

				slog.Error("failed to fetch message", "error", err)
				break
			}

			toProcess <- &m
		}
	}()

	for m := range processed {
		if err := r.CommitMessages(ctx, *m); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
				slog.Debug("context cancelled")
				break
			}
			slog.Error("failed to commit message", "error", err)
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigchan
		slog.Info("termination signal received")
		cancel()
	}()

	run(ctx)
}

func randomString(len int) string {
	b := make([]byte, len)
	for i := range b {
		b[i] = alphabet[rand.Intn(len)]
	}
	return string(b)
}
