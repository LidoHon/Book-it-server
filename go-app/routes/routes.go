package routes

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	AuthRoutes(router)
	BooksRoutes(router)
	WishlistRoutes(router)
	RentRoutes(router)
}
