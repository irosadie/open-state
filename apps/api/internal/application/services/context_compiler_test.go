package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/irosadie/open-state/api/internal/domain/entities"
	domainrag "github.com/irosadie/open-state/api/internal/domain/rag"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
)

// fakeContextRepo is a minimal in-memory IContextRepository for compiler tests.
type fakeContextRepo struct {
	scopes map[string][]entities.ContextRecord
	memory map[string][]entities.MemoryReference
}

func (f *fakeContextRepo) ListContextByScope(_ context.Context, _ string, _ entities.ContextScopeType, scopeID string) ([]entities.ContextRecord, error) {
	return f.scopes[scopeID], nil
}

func (f *fakeContextRepo) ListMemoryByOwner(_ context.Context, _ string, _ string, ownerID string) ([]entities.MemoryReference, error) {
	return f.memory[ownerID], nil
}

// Unused interface methods return no-ops.
func (f *fakeContextRepo) UpsertContext(context.Context, string, entities.ContextScopeType, string, string, []byte, int) (*entities.ContextRecord, error) {
	return nil, nil
}
func (f *fakeContextRepo) FindContextByScope(context.Context, string, entities.ContextScopeType, string, string) (*entities.ContextRecord, error) {
	return nil, nil
}
func (f *fakeContextRepo) DeleteContext(context.Context, string, entities.ContextScopeType, string, string) error {
	return nil
}
func (f *fakeContextRepo) UpsertMemoryReference(context.Context, string, string, string, string, []byte, *string) (*entities.MemoryReference, error) {
	return nil, nil
}
func (f *fakeContextRepo) FindMemoryReference(context.Context, string, string, string, string) (*entities.MemoryReference, error) {
	return nil, nil
}
func (f *fakeContextRepo) DeleteMemoryReference(context.Context, string, string, string, string) error {
	return nil
}

func raw(v string) json.RawMessage {
	return json.RawMessage(v)
}

func TestContextCompilerSplitsAvailableAndMemory(t *testing.T) {
	repo := &fakeContextRepo{
		scopes: map[string][]entities.ContextRecord{
			"conv-1": {
				{Key: "slot.date", Value: raw(`"2026-01-01"`)},
				{Key: "amount", Value: raw(`100`)},
			},
		},
		memory: map[string][]entities.MemoryReference{
			"cust-1": {
				{Name: "preferred_court", Value: raw(`"court A"`)},
			},
		},
	}
	compiler := NewContextCompiler(repo, StubRAGProvider{}, &noopRedactor{})

	got, err := compiler.Compile(context.Background(), CompileArgs{
		TenantID:       "tenant-1",
		ConversationID: "conv-1",
		OwnerType:      "CUSTOMER",
		OwnerID:        "cust-1",
	})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	if got.Available["slot.date"] != "2026-01-01" {
		t.Fatalf("available missing slot.date: %+v", got.Available)
	}
	if got.Available["amount"] != float64(100) {
		t.Fatalf("available amount wrong: %+v", got.Available)
	}
	if got.Memory["preferred_court"] != "court A" {
		t.Fatalf("memory missing: %+v", got.Memory)
	}
}

func TestContextCompilerAppliesRedaction(t *testing.T) {
	repo := &fakeContextRepo{
		memory: map[string][]entities.MemoryReference{
			"cust-1": {
				{Name: "email", Value: raw(`"user@example.com"`)},
			},
		},
	}
	compiler := NewContextCompiler(repo, StubRAGProvider{}, &maskRedactor{})

	got, err := compiler.Compile(context.Background(), CompileArgs{
		TenantID:  "tenant-1",
		OwnerType: "CUSTOMER",
		OwnerID:   "cust-1",
	})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if got.Memory["email"] != "[REDACTED]" {
		t.Fatalf("expected redacted email, got %v", got.Memory["email"])
	}
	if !got.Redacted {
		t.Fatal("expected redacted flag true")
	}
}

// noopRedactor leaves text unchanged.
type noopRedactor struct{}

func (noopRedactor) Redact(_ context.Context, in string) (string, error) { return in, nil }

// maskRedactor replaces every non-empty string with [REDACTED].
type maskRedactor struct{}

func (maskRedactor) Redact(_ context.Context, in string) (string, error) {
	if in == "" {
		return in, nil
	}
	return "[REDACTED]", nil
}

// StubRAGProvider is a test double.
type StubRAGProvider struct{}

func (StubRAGProvider) Retrieve(context.Context, string) (*domainrag.Retrieval, error) {
	return &domainrag.Retrieval{Text: ""}, nil
}

var _ repositories.IContextRepository = (*fakeContextRepo)(nil)
