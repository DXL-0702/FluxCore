package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jaxson/FluxCore/server/config"
)

func NewRouter(cfg config.Config) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(requireAPIToken(cfg.Security.APIToken))
	if err := router.SetTrustedProxies(nil); err != nil {
		panic(err)
	}

	registerRoutes(router, cfg)

	return router
}

func registerRoutes(router *gin.Engine, cfg config.Config) {
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"database": gin.H{
				"type": cfg.Database.Type,
			},
			"service": "fluxcore-server",
			"status":  "ok",
		})
	})
}
