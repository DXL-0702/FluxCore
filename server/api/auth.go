package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const bearerAuthScheme = "Bearer"

func requireAPIToken(expectedToken string) gin.HandlerFunc {
	expectedToken = strings.TrimSpace(expectedToken)

	return func(ctx *gin.Context) {
		if !isAPIPath(ctx.Request.URL.Path) {
			ctx.Next()
			return
		}

		token, ok := bearerTokenFromHeader(ctx.GetHeader("Authorization"))
		if !ok || expectedToken == "" || !secureTokenEqual(token, expectedToken) {
			writeUnauthorized(ctx)
			return
		}

		ctx.Next()
	}
}

func isAPIPath(path string) bool {
	return path == "/api" || strings.HasPrefix(path, "/api/")
}

func bearerTokenFromHeader(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 {
		return "", false
	}
	if !strings.EqualFold(parts[0], bearerAuthScheme) {
		return "", false
	}
	if parts[1] == "" {
		return "", false
	}

	return parts[1], true
}

func secureTokenEqual(actual string, expected string) bool {
	actualHash := sha256.Sum256([]byte(actual))
	expectedHash := sha256.Sum256([]byte(expected))

	return subtle.ConstantTimeCompare(actualHash[:], expectedHash[:]) == 1
}

func writeUnauthorized(ctx *gin.Context) {
	ctx.Header("WWW-Authenticate", `Bearer realm="fluxcore"`)
	ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": gin.H{
			"code":    "unauthorized",
			"message": "authentication required",
		},
	})
}
