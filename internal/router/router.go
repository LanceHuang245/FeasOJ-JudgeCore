package router

import (
	gincontext "main/internal/gin_context"
	"main/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func LoadRouter(r *gin.Engine) *gin.RouterGroup {
	r.Use(middlewares.Logger())
	router1 := r.Group("/api/v1/judgecore")
	{
		router1.GET("/health", gincontext.Health)

		router1.POST("/judge", gincontext.Judge)
	}
	return router1
}
