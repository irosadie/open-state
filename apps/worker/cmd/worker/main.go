package main

import (
	"log"

	"github.com/hibiken/asynq"
	"github.com/irosadie/open-state/worker/internal/application/use-cases"
	"github.com/irosadie/open-state/worker/internal/infrastructure/queue"
)

func main() {
	srv, err := queue.NewServer()
	if err != nil {
		log.Fatalf("worker config error: %v", err)
	}

	// Register job handlers
	mux := asynq.NewServeMux()
	mux.Handle(usecases.TypeRuntimeSummary, usecases.NewRuntimeSummaryHandler())

	log.Println("starting worker...")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("worker error: %v", err)
	}
}
