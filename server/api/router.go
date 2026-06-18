package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	if err := router.SetTrustedProxies(nil); err != nil {
		panic(err)
	}

	registerRoutes(router)

	return router
}

func registerRoutes(router *gin.Engine) {
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"service": "fluxcore-server",
			"status":  "ok",
		})
	})
}
