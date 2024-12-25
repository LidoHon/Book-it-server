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
		userRoutes.POST("/login", controllers.Login())
		userRoutes.POST("/reset-password", controllers.ResetPassword())
		userRoutes.POST("/update-password", controllers.UpdatePassword())
		userRoutes.POST("/delete", controllers.DeleteUser())
		userRoutes.POST("/update-profile", controllers.UpdateProfile())

	}
}
