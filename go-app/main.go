package main

import (
	"fmt"
	"os"

	// "github.com/LidoHon/devConnect/controllers"
	"github.com/LidoHon/devConnect/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(`../.env`)
	if err != nil {
		fmt.Println("error loading enviroment variables", err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	router := gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router)

	// userRoutes := router.Group("/api/users")
	// userRoutes.POST("register", controllers.RegisterUser())

	fmt.Printf("Server running on port %s", port)

	router.Run(":" + port)

}
