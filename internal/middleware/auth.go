package middleware

import (
	"net/http"
	"os"

	"github.com/Flikest/PingVi_backend/pkg/tokens"
	"github.com/gin-gonic/gin"
)

func IsAuthorized(ctx *gin.Context) {
	accessToken := ctx.Request.Header.Get("Authorization")
	if accessToken == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
		return
	}

	payload, err := tokens.Verify(accessToken, []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired access token"})
		return
	}

	ctx.Set(payload.ID, payload)

	ctx.Next()
}
