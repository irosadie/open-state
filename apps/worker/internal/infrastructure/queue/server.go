package queue

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hibiken/asynq"
)

// NewServer creates and configures an asynq server from REDIS_URL.
func NewServer() (*asynq.Server, error) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL is required")
	}

	opt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_URL: %w", err)
	}

	srv := asynq.NewServer(opt, asynq.Config{
		Concurrency: 10,
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			log.Printf("task %s failed: %v", task.Type(), err)
		}),
	})

	return srv, nil
}
