package routes

import (
	"github.com/LidoHon/devConnect/controllers"
	"github.com/LidoHon/devConnect/middlewares"
	"github.com/gin-gonic/gin"
)

func BooksRoutes(incomingRoutes *gin.Engine) {
	booksRoutes := incomingRoutes.Group("/api/books")
	{
		booksRoutes.POST("/insert", middlewares.ImageUpload(), controllers.InsertBooks())
		booksRoutes.GET("/get-books", controllers.GetBooks())
		booksRoutes.GET("/get-books/:id", controllers.GetBooksById())
		// booksRoutes.PUT("/update-books/:id", middlewares.ImageUpload(), controllers.UpdateBooks())
		booksRoutes.PUT("/update-books", middlewares.ImageUpload(), controllers.UpdateBooks())
		booksRoutes.DELETE("/delete-books/:id", controllers.DeleteBooks())
	}
}
