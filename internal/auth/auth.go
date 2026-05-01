package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type JWTManager interface {
	GenerateToken(claims JWTClaims) (string, error)
	ValidateToken(token string) (*JWTClaims, error)
}

type jwtManager struct {
	secretKey     []byte
	tokenDuration time.Duration
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) JWTManager {
	return &jwtManager{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
	}
}

func (m *jwtManager) GenerateToken(claims JWTClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": claims.UserID,
		"role":    claims.Role,
		"exp":     time.Now().Add(m.tokenDuration).Unix(),
		"iat":     time.Now().Unix(),
	})

	return token.SignedString(m.secretKey)
}

func (m *jwtManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return &JWTClaims{
		UserID: claims["user_id"].(string),
		Role:   claims["role"].(string),
	}, nil
}