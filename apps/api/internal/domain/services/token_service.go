package services

// TokenService defines the contract for JWT token operations.
type TokenService interface {
	GenerateToken(userID string) (string, error)
	ValidateToken(token string) (userID string, err error)
	HashToken(token string) string
}
