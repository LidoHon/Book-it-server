package main

import (
	"fmt"
	"github.com/LidoHon/devConnect/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"
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

	routes.RegisterRoutes(router)

	fmt.Printf("Server running on port %s", port)

	router.Run(":" + port)

}
