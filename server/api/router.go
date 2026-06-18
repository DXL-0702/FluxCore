package api

import "github.com/gin-gonic/gin"

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
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"service": "fluxcore-server",
			"status":  "ok",
		})
	})
}
