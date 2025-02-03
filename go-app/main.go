package main

import (
	"fmt"
	"log"
	"os"

	"github.com/LidoHon/devConnect/routes"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

var store *sessions.CookieStore

func main() {
	err := godotenv.Load(`../.env`)
	if err != nil {
		fmt.Println("error loading enviroment variables", err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	// Load session secret from environment variable
	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		log.Fatal("SESSION_SECRET is not set in environment variables")
		sessionSecret = "secret"
	}

	// Assign the session store to `gothic` using the session secret
	store = sessions.NewCookieStore([]byte(sessionSecret))
	gothic.Store = store

	router := gin.New()
	router.Use(gin.Logger())

	// Configure Google OAuth
	goth.UseProviders(
		google.New(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			"http://localhost:"+port+"/auth/google/callback",
			"email", "profile",
		),
	)

	routes.RegisterRoutes(router)

	fmt.Printf("Server running on port %s", port)

	router.Run(":" + port)

}
