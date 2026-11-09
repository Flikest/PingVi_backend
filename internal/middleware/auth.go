package middleware

import (
	"net/http"
	"os"
	"time"

	"github.com/Flikest/PingVi_backend/pkg/tokens"
	"github.com/gin-gonic/gin"
)

const CtxUserPayload = "user_payload"

func IsAuthorized(ctx *gin.Context) {
	accessToken := ctx.Request.Header.Get("Authorization")
	if accessToken == "" {
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid access token"})
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "server misconfigured"})
		return
	}

	payload, err := tokens.Verify(accessToken, []byte(secret))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired access token"})
		return
	}
	if payload == nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token payload"})
		return
	}

	if payload.ExpiresAt == nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token missing exp claim"})
		return
	}

	now := time.Now()
	if !payload.ExpiresAt.After(now) {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
		return
	}

	ctx.Set(CtxUserPayload, payload)
	ctx.Next()
}
