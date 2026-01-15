package api

import (
	"github.com/gin-gonic/gin"

	"judge-service/internal/config"
)

// SetupRouter sets up the Gin router
func SetupRouter(cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	handler := NewHandler(cfg)

	api := r.Group("/api")
	{
		api.POST("/judge", handler.Judge)
		api.GET("/health", handler.Health)
	}

	return r
}
