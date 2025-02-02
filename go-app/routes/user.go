package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/LidoHon/devConnect/controllers"
	"github.com/LidoHon/devConnect/helpers"
	"github.com/LidoHon/devConnect/middlewares"
	// "github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
)

func AuthRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/auth/google", func(c *gin.Context) {
		// session := sessions.Default(c)
		// gothic.Store = session 
		gothic.BeginAuthHandler(c.Writer, c.Request)
	})

	incomingRoutes.GET("/auth/google/callback", func(c *gin.Context) {
		// session := sessions.Default(c)
		// gothic.Store = session
		user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
		if err != nil {
			responseError := map[string]string{
				"message": "Error logging in with google",
				"details": err.Error(),
			}
			jsonData, _ := json.Marshal(responseError)
			escapedData := url.QueryEscape(string(jsonData))
			clientLoginUrl := os.Getenv("CLIENT_LOGIN_URL")
			redirectUrl := clientLoginUrl + "?error=" + escapedData
			c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
			return
		}

		token, refreshToken, userId, userRole, err := helpers.HandleAuth(user.Email, user.Name, user.AvatarURL, user.UserID )
		if err != nil {
			log.Println("Something went wrong:", err.Error())
			responseError := map[string]string{
				"message": "Error logging in with google",
				"detail":  err.Error(),
			}
			jsonData, _ := json.Marshal(responseError)
			escapedData := url.QueryEscape(string(jsonData))
			clientLoginUrl := os.Getenv("CLIENT_LOGIN_URL")
			redirectUrl := clientLoginUrl + "?error=" + escapedData
			c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
			return
		}

		responseData := map[string]string{
			"message":      "user registerd sucessfully",
			"token":        token,
			"refreshToken": refreshToken,
			"id":           strconv.Itoa(userId),
			"role":         userRole,
		}
		jsonData, _ := json.Marshal(responseData)
		escapedData := url.QueryEscape(string(jsonData))
		redirectUrl := os.Getenv("CLIENT_HOMEPAGE_URL") + "?data=" + escapedData
		log.Println("Redirecting to:", redirectUrl)
		c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
	})

	userRoutes := incomingRoutes.Group("/api/users")
	{
		userRoutes.POST("/register", middlewares.ImageUpload(), controllers.RegisterUser())
		userRoutes.POST("/verify-email", controllers.VerifyEmail())
		userRoutes.POST("/login", controllers.Login())
		userRoutes.POST("/reset-password", controllers.ResetPassword())
		userRoutes.POST("/update-password", controllers.UpdatePassword())
		userRoutes.POST("/delete", controllers.DeleteUser())
		userRoutes.POST("/update-profile", middlewares.ImageUpload(), controllers.UpdateProfile())

	}
}
