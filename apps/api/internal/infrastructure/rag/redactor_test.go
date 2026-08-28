package rag

import (
	"context"
	"testing"
)

func TestDefaultRedactorMasksEmail(t *testing.T) {
	r := NewDefaultRedactor()
	out, err := r.Redact(context.Background(), "contact user@example.com please")
	if err != nil {
		t.Fatalf("redact: %v", err)
	}
	if out != "contact [email redacted] please" {
		t.Fatalf("expected email masked, got %q", out)
	}
}

func TestDefaultRedactorMasksPhone(t *testing.T) {
	r := NewDefaultRedactor()
	out, err := r.Redact(context.Background(), "call 0812 3456 7890 now")
	if err != nil {
		t.Fatalf("redact: %v", err)
	}
	if out != "call [phone redacted] now" {
		t.Fatalf("expected phone masked, got %q", out)
	}
}

func TestDefaultRedactorLeavesPlainText(t *testing.T) {
	r := NewDefaultRedactor()
	out, err := r.Redact(context.Background(), "book a padel court")
	if err != nil {
		t.Fatalf("redact: %v", err)
	}
	if out != "book a padel court" {
		t.Fatalf("expected unchanged, got %q", out)
	}
}

func TestStubRAGProviderReturnsEmpty(t *testing.T) {
	p := StubRAGProvider{}
	r, err := p.Retrieve(context.Background(), "anything")
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}
	if r.Text != "" {
		t.Fatalf("expected empty retrieval, got %q", r.Text)
	}
}
