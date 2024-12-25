package routes

import (
	"github.com/LidoHon/devConnect/controllers"
	"github.com/gin-gonic/gin"
)

func BooksRoutes(incomingRoutes *gin.Engine) {
	booksRoutes := incomingRoutes.Group("/api/books")
	{
		booksRoutes.POST("/insert", controllers.InsertBooks())
		booksRoutes.GET("/get-books", controllers.GetBooks())
		booksRoutes.GET("/get-books/:id", controllers.GetBooksById())
		booksRoutes.PUT("/update-books/:id", controllers.UpdateBooks())
		booksRoutes.DELETE("/delete-books/:id", controllers.DeleteBooks())
	}
}
