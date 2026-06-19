package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jaxson/FluxCore/server/config"
	"github.com/jaxson/FluxCore/server/service"
	"gorm.io/gorm"
)

func NewRouter(cfg config.Config, conn *gorm.DB) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	if err := router.SetTrustedProxies(nil); err != nil {
		panic(err)
	}

	registerRoutes(router, cfg, service.NewProjectService(conn))

	return router
}

func registerRoutes(router *gin.Engine, cfg config.Config, projects *service.ProjectService) {
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"database": gin.H{
				"type": cfg.Database.Type,
			},
			"service": "fluxcore-server",
			"status":  "ok",
		})
	})

	api := router.Group("/api")
	api.Use(requireAPIToken(cfg.Security.APIToken))
	registerProjectRoutes(api, projects)

	router.NoRoute(func(ctx *gin.Context) {
		if isAPIPath(ctx.Request.URL.Path) && !isAuthorized(ctx.GetHeader("Authorization"), cfg.Security.APIToken) {
			writeUnauthorized(ctx)
			return
		}

		writeNotFound(ctx)
	})
}
