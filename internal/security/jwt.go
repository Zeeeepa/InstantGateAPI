package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/proyaai/instantgate/internal/config"
)

type JWTManager struct {
	secretKey []byte
	expiry    time.Duration
	issuer    string
}

type Claims struct {
	jwt.RegisteredClaims
	UserID    string   `json:"uid,omitempty"`
	Username  string   `json:"username,omitempty"`
	Roles     []string `json:"roles,omitempty"`
	TokenType string   `json:"type,omitempty"`
}

func NewJWTManager(cfg *config.JWTConfig) *JWTManager {
	return &JWTManager{
		secretKey: []byte(cfg.Secret),
		expiry:    cfg.Expiry,
		issuer:    cfg.Issuer,
	}
}

func (j *JWTManager) GenerateToken(userID, username string, roles []string) (string, error) {
	now := time.Now()
	tokenID := uuid.New().String()

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.expiry)),
			NotBefore: jwt.NewNumericDate(now),
			ID:        tokenID,
		},
		UserID:    userID,
		Username:  username,
		Roles:     roles,
		TokenType: "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		}
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	if claims.Issuer != j.issuer {
		return nil, ErrInvalidIssuer
	}

	return claims, nil
}

func (j *JWTManager) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	return j.GenerateToken(claims.UserID, claims.Username, claims.Roles)
}

var (
	ErrTokenExpired   = errors.New("token has expired")
	ErrTokenMalformed = errors.New("token is malformed")
	ErrTokenInvalid   = errors.New("token is invalid")
	ErrInvalidIssuer  = errors.New("invalid token issuer")
	ErrNoToken        = errors.New("no token provided")
)
