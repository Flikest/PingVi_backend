package tokens

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type CustomClaimsAccess struct {
	ID uuid.UUID
	jwt.RegisteredClaims
}
type CustomClaims2FA struct {
	ID     uuid.UUID
	Locale string
	Device string
	jwt.RegisteredClaims
}

func CreateAccessToken(userID uuid.UUID, secret []byte) (string, error) {
	expTime := time.Now().Add(time.Minute * 15)

	claims := &CustomClaimsAccess{
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

func CreateSessionToken(bytesLength int) (string, error) {
	tokenBytes := make([]byte, bytesLength)

	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", fmt.Errorf("critical error generating random bytes: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	return token, nil
}

func Create2FaToken(userID uuid.UUID, locale string, device string, secret []byte) (string, error) {
	expTime := time.Now().Add(time.Minute * 5)

	claims := &CustomClaims2FA{
		ID:     userID,
		Locale: locale,
		Device: device,
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

func GenerateTemporaryToken() (string, error) {
	tokenBytes := make([]byte, 32)

	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", fmt.Errorf("critical error generating random bytes: %w", err)
	}

	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	return token, nil
}

func Verify(tokenString string, secret []byte) (*CustomClaimsAccess, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaimsAccess{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Invalid signature method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("token parsing error: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaimsAccess)
	if !ok || claims == nil {
		return nil, fmt.Errorf("unable to convert token statements")
	}

	return claims, nil
}

func Verify2FA(tokenString string, secret []byte) (*CustomClaims2FA, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims2FA{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Invalid signature method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("token parsing error: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims2FA)
	if !ok || claims == nil {
		return nil, fmt.Errorf("unable to convert token statements")
	}

	return claims, nil
}
