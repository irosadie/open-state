package usecases

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/hibiken/asynq"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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

	// Job span (PRD §84). No-op without a global TracerProvider.
	_, span := otel.Tracer("openstate.worker").Start(ctx, "process job")
	defer span.End()
	span.SetAttributes(attribute.String("task.type", task.Type()))

	slog.Info("processing runtime summary", "task", task.Type(), "message", payload.Message)
	return nil
}
