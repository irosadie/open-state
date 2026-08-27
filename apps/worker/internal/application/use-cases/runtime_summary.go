package usecases

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
)

const TypeRuntimeSummary = "runtime:summary"

type RuntimeSummaryPayload struct {
	Message string `json:"message"`
}

// RuntimeSummaryHandler handles runtime summary jobs.
type RuntimeSummaryHandler struct{}

func NewRuntimeSummaryHandler() *RuntimeSummaryHandler {
	return &RuntimeSummaryHandler{}
}

func (h *RuntimeSummaryHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	var payload RuntimeSummaryPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return err
	}

	log.Printf("[runtime:summary] processing: %s", payload.Message)
	return nil
}
