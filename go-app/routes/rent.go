package routes

import (
	"github.com/LidoHon/devConnect/controllers"
	"github.com/gin-gonic/gin"
)

func RentRoutes(incomingRoutes *gin.Engine) {
	rentRoutes := incomingRoutes.Group("/api/rent")
	{
		rentRoutes.POST("/create", controllers.PlaceRent())
		rentRoutes.DELETE("/delete/:id", controllers.DeleteRent())
		rentRoutes.PUT("/return-book", controllers.ReturnBook())

	}
}
