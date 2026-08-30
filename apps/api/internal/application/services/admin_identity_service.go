package services

import (
	"context"
	"strings"

	"github.com/irosadie/open-state/api/internal/application/dtos"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"github.com/irosadie/open-state/api/internal/domain/repositories"
	domain "github.com/irosadie/open-state/go-shared/domain"
)

const (
	defaultMembershipPageSize = 20
	maxMembershipPageSize     = 100
)

// AdminIdentityService owns current-tenant profile and membership policy. The
// tenant id is always supplied by the authenticated request context.
type AdminIdentityService struct {
	repo  repositories.IAdminRepository
	audit *AuditWriter
}

func NewAdminIdentityService(repo repositories.IAdminRepository, audit *AuditWriter) *AdminIdentityService {
	return &AdminIdentityService{repo: repo, audit: audit}
}

func (s *AdminIdentityService) GetTenant(ctx context.Context, tenantID string) (*dtos.TenantDTO, error) {
	tenant, err := s.repo.FindTenantByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return toTenantDTO(tenant), nil
}

func (s *AdminIdentityService) UpdateTenant(ctx context.Context, tenantID, actor string, input dtos.UpdateTenantRequest, correlationID *string) (*dtos.TenantDTO, error) {
	current, err := s.repo.FindTenantByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	name := current.Name
	if input.Name != nil {
		name = strings.TrimSpace(*input.Name)
	}
	slug := current.Slug
	if input.Slug != nil {
		slug = strings.TrimSpace(*input.Slug)
	}
	description := current.Description
	if input.Description != nil {
		description = strings.TrimSpace(*input.Description)
	}
	if name == "" || slug == "" {
		err := domain.NewValidation("tenant name and slug are required")
		s.writeAudit(ctx, tenantID, actor, entities.AuditActionTenantUpdated, tenantID,
			toTenantAudit(current), map[string]any{"outcome": "rejected", "error": err.Error()}, correlationID)
		return nil, err
	}
	if !validTenantSlug(slug) {
		err := domain.NewValidation("tenant slug is invalid")
		s.writeAudit(ctx, tenantID, actor, entities.AuditActionTenantUpdated, tenantID,
			toTenantAudit(current), map[string]any{"outcome": "rejected", "error": err.Error()}, correlationID)
		return nil, err
	}

	updated, err := s.repo.UpdateTenantProfile(ctx, tenantID, name, slug, description)
	if err != nil {
		s.writeAudit(ctx, tenantID, actor, entities.AuditActionTenantUpdated, tenantID,
			toTenantAudit(current), map[string]any{"outcome": "rejected", "error": err.Error()}, correlationID)
		return nil, err
	}
	s.writeAudit(ctx, tenantID, actor, entities.AuditActionTenantUpdated, tenantID,
		toTenantAudit(current), toTenantAudit(updated), correlationID)
	return toTenantDTO(updated), nil
}

func validTenantSlug(slug string) bool {
	if len(slug) == 0 || len(slug) > 255 {
		return false
	}
	for _, char := range slug {
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '-' {
			return false
		}
	}
	return true
}

func (s *AdminIdentityService) ListMemberships(ctx context.Context, tenantID, search string, page, pageSize int) (*dtos.TenantMembershipPageDTO, error) {
	page, pageSize = normalizePage(page, pageSize)
	var searchFilter *string
	if value := strings.TrimSpace(search); value != "" {
		searchFilter = &value
	}

	total, err := s.repo.CountMemberships(ctx, tenantID, searchFilter)
	if err != nil {
		return nil, err
	}
	members, err := s.repo.ListMemberships(ctx, tenantID, searchFilter, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	data := make([]dtos.TenantMembershipDTO, 0, len(members))
	for i := range members {
		data = append(data, toMembershipDTO(&members[i]))
	}
	return &dtos.TenantMembershipPageDTO{
		Data:     data,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		HasNext:  int64(page*pageSize) < total,
	}, nil
}

func (s *AdminIdentityService) UpdateMembershipRole(ctx context.Context, tenantID, actor, userID, roleValue string, correlationID *string) (*dtos.TenantMembershipDTO, error) {
	role := entities.UserRole(strings.ToUpper(strings.TrimSpace(roleValue)))
	if !entities.IsTenantRole(role) {
		err := domain.NewValidation("invalid tenant role")
		s.writeAudit(ctx, tenantID, actor, entities.AuditActionRoleUpdated, userID,
			nil, map[string]any{"outcome": "rejected", "error": err.Error(), "requestedRole": role}, correlationID)
		return nil, err
	}

	var before *entities.TenantMembership
	var updated *entities.TenantMembership
	err := s.repo.WithTx(ctx, func(tx repositories.IAdminRepository) error {
		member, err := tx.FindMembership(ctx, tenantID, userID)
		if err != nil {
			return err
		}
		before = member
		if member.Role == role {
			updated = member
			return nil
		}
		if member.Role == entities.UserRoleOwner && role != entities.UserRoleOwner {
			owners, err := tx.CountOwners(ctx, tenantID)
			if err != nil {
				return err
			}
			if owners <= 1 {
				return domain.NewConflict("tenant must retain at least one Owner")
			}
		}
		updated, err = tx.AssignMembershipRole(ctx, tenantID, userID, role)
		return err
	})
	if err != nil {
		var beforeAudit any
		if before != nil {
			beforeAudit = toMembershipAudit(before)
		}
		s.writeAudit(ctx, tenantID, actor, entities.AuditActionRoleUpdated, userID,
			beforeAudit, map[string]any{"outcome": "rejected", "error": err.Error(), "requestedRole": role}, correlationID)
		return nil, err
	}

	s.writeAudit(ctx, tenantID, actor, entities.AuditActionRoleUpdated, userID,
		toMembershipAudit(before), toMembershipAudit(updated), correlationID)
	return membershipDTOOrNil(updated), nil
}

func (s *AdminIdentityService) RemoveMembership(ctx context.Context, tenantID, actor, userID string, correlationID *string) error {
	var before *entities.TenantMembership
	err := s.repo.WithTx(ctx, func(tx repositories.IAdminRepository) error {
		member, err := tx.FindMembership(ctx, tenantID, userID)
		if err != nil {
			return err
		}
		before = member
		if member.Role == entities.UserRoleOwner {
			owners, err := tx.CountOwners(ctx, tenantID)
			if err != nil {
				return err
			}
			if owners <= 1 {
				return domain.NewConflict("tenant must retain at least one Owner")
			}
		}
		return tx.RemoveMembership(ctx, tenantID, userID)
	})
	if err != nil {
		var beforeAudit any
		if before != nil {
			beforeAudit = toMembershipAudit(before)
		}
		s.writeAudit(ctx, tenantID, actor, entities.AuditActionRoleRemoved, userID,
			beforeAudit, map[string]any{"outcome": "rejected", "error": err.Error()}, correlationID)
		return err
	}
	s.writeAudit(ctx, tenantID, actor, entities.AuditActionRoleRemoved, userID,
		toMembershipAudit(before), map[string]any{"outcome": "removed"}, correlationID)
	return nil
}

func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultMembershipPageSize
	}
	if pageSize > maxMembershipPageSize {
		pageSize = maxMembershipPageSize
	}
	return page, pageSize
}

func toTenantDTO(tenant *entities.Tenant) *dtos.TenantDTO {
	return &dtos.TenantDTO{
		ID:          tenant.ID,
		Name:        tenant.Name,
		Slug:        tenant.Slug,
		Description: tenant.Description,
		CreatedAt:   tenant.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   tenant.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toMembershipDTO(member *entities.TenantMembership) dtos.TenantMembershipDTO {
	return dtos.TenantMembershipDTO{
		RoleAssignmentID: member.RoleAssignmentID,
		UserID:           member.UserID,
		TenantID:         member.TenantID,
		Role:             string(member.Role),
		Email:            member.Email,
		Name:             member.Name,
		Status:           string(member.Status),
		Photo:            member.Photo,
		CreatedAt:        member.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        member.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func membershipDTOOrNil(member *entities.TenantMembership) *dtos.TenantMembershipDTO {
	if member == nil {
		return nil
	}
	dto := toMembershipDTO(member)
	return &dto
}

func toTenantAudit(tenant *entities.Tenant) map[string]any {
	return map[string]any{"id": tenant.ID, "name": tenant.Name, "slug": tenant.Slug, "description": tenant.Description}
}

func toMembershipAudit(member *entities.TenantMembership) map[string]any {
	if member == nil {
		return nil
	}
	return map[string]any{"userId": member.UserID, "role": member.Role, "email": member.Email}
}

func (s *AdminIdentityService) writeAudit(ctx context.Context, tenantID, actor string, action entities.AuditAction, resourceID string, before, after any, correlationID *string) {
	if s.audit != nil {
		resourceType := "tenant"
		if action == entities.AuditActionRoleUpdated || action == entities.AuditActionRoleRemoved {
			resourceType = "role_assignment"
		}
		s.audit.Write(ctx, tenantID, actor, action, resourceType, resourceID, before, after, correlationID)
	}
}
