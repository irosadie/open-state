package services

import (
	"context"
	"testing"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
)

type recorderTraceRepo struct {
	entries []entities.RuntimeTraceEntry
}

func (f *recorderTraceRepo) Append(_ context.Context, tenantID string, input repositories.AppendRuntimeTraceInput) (*entities.RuntimeTraceEntry, error) {
	entry := entities.RuntimeTraceEntry{
		ID:                 "trace-" + string(rune(len(f.entries)+1)),
		TenantID:           tenantID,
		WorkflowInstanceID: input.WorkflowInstanceID,
		Sequence:           int64(len(f.entries) + 1),
		TurnID:             input.TurnID,
		Stage:              input.Stage,
		Source:             input.Source,
		Status:             input.Status,
		OccurredAt:         input.OccurredAt,
		CorrelationID:      input.CorrelationID,
		DurationMS:         input.DurationMS,
		ReasonCode:         input.ReasonCode,
		ErrorCode:          input.ErrorCode,
		ProviderAlias:      input.ProviderAlias,
		ProviderReference:  input.ProviderReference,
		Summary:            input.Summary,
		Attributes:         input.Attributes,
	}
	f.entries = append(f.entries, entry)
	return &f.entries[len(f.entries)-1], nil
}

func (f *recorderTraceRepo) ListByInstance(context.Context, string, string) ([]entities.RuntimeTraceEntry, error) {
	return f.entries, nil
}

func (f *recorderTraceRepo) ListByTurn(context.Context, string, string, string) ([]entities.RuntimeTraceEntry, error) {
	return f.entries, nil
}

func TestRuntimeTraceRecorderSanitizesBeforePersistence(t *testing.T) {
	repo := &recorderTraceRepo{}
	recorder := NewRuntimeTraceRecorder(repo)
	_, err := recorder.RecordExternal(context.Background(), "tenant-a", "instance-a", entities.RuntimeTraceStageRAGIntegration, ExternalProviderMetadata{
		ProviderAlias:      "knowledge-search",
		OperationReference: "op-123",
		Status:             entities.RuntimeTraceStatusSucceeded,
		Summary:            "retrieved private document for user@example.com",
		Attributes: map[string]any{
			"raw_response":        "should not persist",
			"result_count":        3,
			"retrieved_documents": []any{"private"},
		},
	})
	if err != nil {
		t.Fatalf("record external: %v", err)
	}
	if len(repo.entries) != 1 {
		t.Fatalf("expected one entry, got %d", len(repo.entries))
	}
	entry := repo.entries[0]
	if entry.Source != entities.RuntimeTraceSourceExternalProvider {
		t.Fatalf("expected external source, got %s", entry.Source)
	}
	if entry.Attributes["raw_response"] != "[REDACTED]" {
		t.Fatalf("raw response was not redacted: %#v", entry.Attributes)
	}
	if entry.Attributes["retrieved_documents"] != "[REDACTED]" {
		t.Fatalf("retrieved documents were not redacted: %#v", entry.Attributes)
	}
	if *entry.Summary == "retrieved private document for user@example.com" {
		t.Fatal("raw provider summary was retained")
	}
}
