package routes

import (
	"github.com/LidoHon/devConnect/controllers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine) {
	userRoutes := incomingRoutes.Group("/api/users")
	{
		// You can add more routes as needed
		userRoutes.POST("/register", controllers.RegisterUser())
		userRoutes.POST("/verify-email", controllers.VerifyEmail())
		// userRoutes.GET("/verify-email", controllers.VerifyEmail())

	}
}
