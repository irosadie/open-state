package infrastructure

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/irosadie/open-state/api/internal/domain/services"
	"golang.org/x/crypto/sha3"
	"encoding/hex"
)

type JwtTokenService struct {
	secret []byte
	ttl    time.Duration
}

func NewJwtTokenService(secret string) services.TokenService {
	return &JwtTokenService{
		secret: []byte(secret),
		ttl:    24 * time.Hour,
	}
}

func (s *JwtTokenService) GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(s.ttl).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JwtTokenService) ValidateToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("invalid subject")
	}

	return sub, nil
}

func (s *JwtTokenService) HashToken(token string) string {
	h := sha3.New256()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}
