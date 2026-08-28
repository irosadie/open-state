package services

import (
	"context"
	"testing"

	"github.com/irosadie/open-state/api/internal/domain/entities"
)

type stubRoleRepo struct {
	roles map[string]entities.UserRole // key: userID+":"+tenantID
}

func (s *stubRoleRepo) FindRoleByUserAndTenant(_ context.Context, userID, tenantID string) (entities.UserRole, error) {
	if r, ok := s.roles[userID+":"+tenantID]; ok {
		return r, nil
	}
	return "", nil
}

func TestAuthorizationService(t *testing.T) {
	svc := NewAuthorizationService(&stubRoleRepo{roles: map[string]entities.UserRole{
		"u1:t1": entities.UserRoleAdmin,
		"u1:t2": entities.UserRoleViewer,
	}})
	ctx := context.Background()

	// Admin in t1 can publish.
	if err := svc.Require(ctx, "u1", "t1", "workflow:publish"); err != nil {
		t.Errorf("expected admin u1/t1 authorized, got %v", err)
	}
	// Viewer in t2 cannot publish.
	if err := svc.Require(ctx, "u1", "t2", "workflow:publish"); err == nil {
		t.Error("expected viewer u1/t2 to be denied workflow:publish")
	}
	// Absent assignment defaults to VIEWER -> denied elevated permission.
	if err := svc.Require(ctx, "u9", "t9", "workflow:publish"); err == nil {
		t.Error("expected absent role assignment to default-deny elevated permission")
	}
	// Absent assignment defaults to VIEWER -> allowed read.
	if err := svc.Require(ctx, "u9", "t9", "workflow:read"); err != nil {
		t.Errorf("expected absent assignment to allow read (VIEWER), got %v", err)
	}
}
