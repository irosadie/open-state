package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/irosadie/open-state/go-shared/domain"
	"github.com/irosadie/open-state/api/internal/application/use-cases"
	"github.com/irosadie/open-state/api/internal/domain/entities"
	"golang.org/x/crypto/bcrypt"
)

// --- Mocks ---

type mockAuthRepo struct {
	users    map[string]*entities.User
	sessions map[string]*entities.AuthSession
}

func newMockRepo() *mockAuthRepo {
	return &mockAuthRepo{
		users:    make(map[string]*entities.User),
		sessions: make(map[string]*entities.AuthSession),
	}
}

func (m *mockAuthRepo) FindUserByEmail(_ context.Context, email string) (*entities.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockAuthRepo) FindUserByID(_ context.Context, id string) (*entities.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (m *mockAuthRepo) CreateUser(_ context.Context, email, passwordHash, name string, role entities.UserRole, status entities.UserStatus) (*entities.User, error) {
	u := &entities.User{
		ID:           "user-1",
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
		Role:         role,
		Status:       status,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	m.users[u.ID] = u
	return u, nil
}

func (m *mockAuthRepo) CreateSession(_ context.Context, userID, tokenHash string, expiresAt time.Time) (*entities.AuthSession, error) {
	s := &entities.AuthSession{
		ID:        "session-1",
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	m.sessions[tokenHash] = s
	return s, nil
}

func (m *mockAuthRepo) FindSessionByTokenHash(_ context.Context, tokenHash string) (*entities.AuthSession, error) {
	s, ok := m.sessions[tokenHash]
	if !ok {
		return nil, errors.New("not found")
	}
	return s, nil
}

func (m *mockAuthRepo) DeleteSessionByTokenHash(_ context.Context, tokenHash string) error {
	delete(m.sessions, tokenHash)
	return nil
}

type mockTokenSvc struct{}

func (m *mockTokenSvc) GenerateToken(userID string) (string, error) {
	return "token-" + userID, nil
}

func (m *mockTokenSvc) ValidateToken(token string) (string, error) {
	if len(token) < 7 {
		return "", errors.New("invalid")
	}
	return token[6:], nil
}

func (m *mockTokenSvc) HashToken(token string) string {
	return "hash-" + token
}

// --- Tests ---

func TestRegisterUser_Success(t *testing.T) {
	repo := newMockRepo()
	uc := usecases.NewRegisterUserUseCase(repo, &mockTokenSvc{})

	user, err := uc.Execute(context.Background(), usecases.RegisterUserInput{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", user.Email)
	}
}

func TestRegisterUser_DuplicateEmail(t *testing.T) {
	repo := newMockRepo()
	uc := usecases.NewRegisterUserUseCase(repo, &mockTokenSvc{})

	_, _ = uc.Execute(context.Background(), usecases.RegisterUserInput{
		Email: "dup@example.com", Password: "password123", Name: "User",
	})

	_, err := uc.Execute(context.Background(), usecases.RegisterUserInput{
		Email: "dup@example.com", Password: "password123", Name: "User2",
	})

	var de *domain.DomainError
	if !errors.As(err, &de) || de.Code != domain.ErrConflict {
		t.Errorf("expected CONFLICT error, got %v", err)
	}
}

func TestRegisterUser_MissingFields(t *testing.T) {
	repo := newMockRepo()
	uc := usecases.NewRegisterUserUseCase(repo, &mockTokenSvc{})

	_, err := uc.Execute(context.Background(), usecases.RegisterUserInput{
		Email: "", Password: "", Name: "",
	})

	var de *domain.DomainError
	if !errors.As(err, &de) || de.Code != domain.ErrValidation {
		t.Errorf("expected VALIDATION error, got %v", err)
	}
}

func TestLoginUser_Success(t *testing.T) {
	repo := newMockRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repo.users["user-1"] = &entities.User{
		ID:           "user-1",
		Email:        "login@example.com",
		PasswordHash: string(hash),
		Name:         "Login User",
		Role:         entities.UserRoleLegacy,
		Status:       entities.UserStatusActive,
	}

	uc := usecases.NewLoginUserUseCase(repo, &mockTokenSvc{})
	out, err := uc.Execute(context.Background(), usecases.LoginUserInput{
		Email:    "login@example.com",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out.AccessToken == "" {
		t.Error("expected access token, got empty string")
	}
}

func TestLoginUser_WrongPassword(t *testing.T) {
	repo := newMockRepo()
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	repo.users["user-1"] = &entities.User{
		ID: "user-1", Email: "u@example.com", PasswordHash: string(hash),
		Role: entities.UserRoleLegacy, Status: entities.UserStatusActive,
	}

	uc := usecases.NewLoginUserUseCase(repo, &mockTokenSvc{})
	_, err := uc.Execute(context.Background(), usecases.LoginUserInput{
		Email: "u@example.com", Password: "wrong",
	})

	var de *domain.DomainError
	if !errors.As(err, &de) || de.Code != domain.ErrUnauthorized {
		t.Errorf("expected UNAUTHORIZED error, got %v", err)
	}
}

func TestLogoutUser_Success(t *testing.T) {
	repo := newMockRepo()
	tokenSvc := &mockTokenSvc{}
	repo.sessions["hash-token-user-1"] = &entities.AuthSession{
		ID: "s1", UserID: "user-1", TokenHash: "hash-token-user-1",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	uc := usecases.NewLogoutUserUseCase(repo, tokenSvc)
	err := uc.Execute(context.Background(), "token-user-1")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, exists := repo.sessions["hash-token-user-1"]; exists {
		t.Error("expected session to be deleted")
	}
}

func TestLogoutUser_SessionNotFound(t *testing.T) {
	repo := newMockRepo()
	uc := usecases.NewLogoutUserUseCase(repo, &mockTokenSvc{})

	err := uc.Execute(context.Background(), "token-nonexistent")

	var de *domain.DomainError
	if !errors.As(err, &de) || de.Code != domain.ErrUnauthorized {
		t.Errorf("expected UNAUTHORIZED error, got %v", err)
	}
}

func TestGetCurrentUser_Success(t *testing.T) {
	repo := newMockRepo()
	repo.users["user-1"] = &entities.User{
		ID: "user-1", Email: "me@example.com", Name: "Me",
		Role: entities.UserRoleLegacy, Status: entities.UserStatusActive,
	}

	uc := usecases.NewGetCurrentUserUseCase(repo)
	user, err := uc.Execute(context.Background(), "user-1")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != "user-1" {
		t.Errorf("expected user-1, got %s", user.ID)
	}
}

func TestGetCurrentUser_NotFound(t *testing.T) {
	repo := newMockRepo()
	uc := usecases.NewGetCurrentUserUseCase(repo)

	_, err := uc.Execute(context.Background(), "nonexistent")

	var de *domain.DomainError
	if !errors.As(err, &de) || de.Code != domain.ErrNotFound {
		t.Errorf("expected NOT_FOUND error, got %v", err)
	}
}
