package tokens

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type CustomClaims struct {
	ID uuid.UUID
	jwt.RegisteredClaims
}

func CreateAccessToken(userID uuid.UUID, secret []byte) (string, error) {
	expTime := time.Now().Add(time.Minute * 15)

	claims := &CustomClaims{
		ID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}

	return tokenString, nil
}

func CreateRefreshToken(bytesLength uint) (string, error) {
	tokenBytes := make([]byte, bytesLength)

	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", fmt.Errorf("critical error generating random bytes: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	return token, nil
}

func Verify(tokenString string, secret []byte) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Invalid signature method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("token parsing error: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || claims == nil {
		return nil, fmt.Errorf("unable to convert token statements")
	}

	return claims, nil
}
