package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func writeAPIError(ctx *gin.Context, status int, code string, message string) {
	ctx.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}

func writeNotFound(ctx *gin.Context) {
	writeAPIError(ctx, http.StatusNotFound, "not_found", "resource not found")
}
