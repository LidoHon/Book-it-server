package routes

import (
	"github.com/LidoHon/devConnect/controllers"
	"github.com/gin-gonic/gin"
)

func WishlistRoutes(incomingRoutes *gin.Engine) {
	WishlistRoutes := incomingRoutes.Group("/api/wishlist")
	{
		WishlistRoutes.POST("/create", controllers.CreateWishlist())
		WishlistRoutes.DELETE("/delete/:id", controllers.DeleteWishList())
	}

}
