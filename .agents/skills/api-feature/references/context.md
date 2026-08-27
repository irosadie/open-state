# Context: API Feature

## Target Folders

```
apps/api/
├── db/
│   ├── migrations/       → goose SQL migration files
│   └── queries/          → sqlc .sql query files ({domain}.sql)
└── internal/
    ├── application/
    │   ├── dtos/         → {domain}.go
    │   └── services/     → {domain}_service.go
    ├── domain/
    │   ├── entities/     → {domain}.go
    │   ├── repositories/ → i_{domain}_repository.go
    │   └── use-cases/    → {verb}_{domain}.go
    ├── infrastructure/
    │   ├── db/           → sqlc generated (DO NOT EDIT)
    │   └── database/     → pgx_{domain}_repository.go
    └── interfaces/http/
        ├── controllers/  → {domain}_controller.go
        ├── middleware/   → error_handler.go, jwt.go, auth_session.go
        └── routes/       → routes.go
```

## Pattern per Layer

### Entity
```go
package entities

import "time"

type User struct {
  ID        string
  Email     string
  Name      string
  Role      UserRole
  Status    UserStatus
  CreatedAt time.Time
  UpdatedAt time.Time
}
```

### Repository Interface
```go
package repositories

import (
  "context"
  "github.com/vibecoding-starter/api/internal/domain/entities"
)

type IUserRepository interface {
  FindByID(ctx context.Context, id string) (*entities.User, error)
  Create(ctx context.Context, email, name string) (*entities.User, error)
}
```

### Use Case
```go
package usecases

import (
  domain "github.com/vibecoding-starter/go-shared/domain"
  "github.com/vibecoding-starter/api/internal/domain/repositories"
)

type CreateUserUseCase struct {
  repo repositories.IUserRepository
}

func NewCreateUserUseCase(repo repositories.IUserRepository) *CreateUserUseCase {
  return &CreateUserUseCase{repo: repo}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, email, name string) (*entities.User, error) {
  existing, _ := uc.repo.FindByEmail(ctx, email)
  if existing != nil {
    return nil, domain.NewConflict("email already in use")
  }
  return uc.repo.Create(ctx, email, name)
}
```

### Service
```go
package services

type UserService struct {
  createUC *usecases.CreateUserUseCase
}

func (s *UserService) Create(ctx context.Context, email, name string) (*dtos.UserDTO, error) {
  user, err := s.createUC.Execute(ctx, email, name)
  if err != nil {
    return nil, err
  }
  return toUserDTO(user), nil
}
```

### Controller
```go
package controllers

type UserController struct {
  svc *services.UserService
}

func (ctrl *UserController) Create(c echo.Context) error {
  var req dtos.CreateUserRequest
  if err := c.Bind(&req); err != nil {
    return err
  }
  result, err := ctrl.svc.Create(c.Request().Context(), req.Email, req.Name)
  if err != nil {
    return err
  }
  return c.JSON(http.StatusCreated, map[string]interface{}{"data": result})
}
```

### Route
```go
func RegisterUserRoutes(e *echo.Echo, ctrl *controllers.UserController, tokenSvc services.TokenService) {
  g := e.Group("/api/users", middleware.JWT(tokenSvc))
  g.POST("/", ctrl.Create)
  g.GET("/:id", ctrl.GetByID)
}
```

### sqlc Query File
```sql
-- name: FindUserByID :one
SELECT id, email, name, role, status, created_at, updated_at
FROM users
WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (email, name, role, status)
VALUES ($1, $2, $3, $4)
RETURNING id, email, name, role, status, created_at, updated_at;
```
