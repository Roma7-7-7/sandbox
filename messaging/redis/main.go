package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

var rdc *redis.Client

type Job struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func publish(ctx context.Context, rdc *redis.Client, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		job := Job{X: rand.Intn(100), Y: rand.Intn(100)}
		body, err := json.Marshal(job)
		if err != nil {
			slog.Warn("failed to marshal job", "error", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if err = rdc.Publish(ctx, "jobs", body).Err(); err != nil {
			if errors.Is(err, context.Canceled) {
				slog.Info("context cancelled, stopping publisher")
				break
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func subscribe(ctx context.Context, rdc *redis.Client, wg *sync.WaitGroup) {
	defer wg.Done()

	subscriber := rdc.Subscribe(ctx, "jobs")
	defer subscriber.Close()

	go func() {
		<-ctx.Done()
		slog.Info("context cancelled, stopping subscriber")
		subscriber.Close()
	}()

	for message := range subscriber.Channel() {
		var job Job
		if err := json.Unmarshal([]byte(message.Payload), &job); err != nil {
			slog.Warn("failed to unmarshal job", "error", err)
			continue
		}

		slog.Debug("received job", "job", job)
		slog.Info("processing job", "x", job.X, "y", job.Y, "result", job.X+job.Y)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rdc = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	defer rdc.Close()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go publish(ctx, rdc, wg)
	wg.Add(1)
	go subscribe(ctx, rdc, wg)

	go func() {
		<-sigchan
		slog.Info("termination signal received")
		cancel()
	}()

	wg.Wait()
	slog.Info("exiting")
}
