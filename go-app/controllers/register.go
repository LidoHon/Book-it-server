package controllers

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/LidoHon/devConnect/helpers"
	"github.com/LidoHon/devConnect/libs"
	"github.com/LidoHon/devConnect/models"
	"github.com/LidoHon/devConnect/requests"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/shurcooL/graphql"
)

var validate = validator.New()

func RegisterUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()
		_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Declare a variable to hold the input data
		var request requests.RegisterRequest

		// Bind the JSON request body to the input struct
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input data"})
			return
		}

		// Validate the input data
		if err := validate.Struct(request); err != nil {
			c.JSON(400, gin.H{"error": "Validation failed", "details": err.Error()})
			return
		}

		var query struct {
			User []struct {
				ID       graphql.Int    `graphql:"id"`
				Name     graphql.String `graphql:"username"`
				Email    graphql.String `graphql:"email"`
				Password graphql.String `graphql:"password"`
				Role     graphql.String `graphql:"role"`
			} `graphql:"users(where: {email: {_eq: $email}})"`
		}

		variable := map[string]interface{}{
			"email": graphql.String(request.Input.Email),
		}

		err := client.Query(context.Background(), &query, variable)
		if err != nil {
			log.Println("Error querying existing user:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
			return
		}

		if len(query.User) != 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
			return

		}

		imageUrl, exists := c.Get("imageUrl")
		if !exists {
			imageUrl = ""
		}

		profilePictureURL, ok := imageUrl.(string)
		if !ok {
			profilePictureURL = ""
		}

		password := helpers.HashedPassword(request.Input.Password)
		var mutation struct {
			CreateUser struct {
				ID       graphql.Int    `graphql:"id"`
				UserName graphql.String `graphql:"username"`
				Email    graphql.String `graphql:"email"`
				Profile  graphql.String `graphql:"profile"`
				Role     graphql.String `graphql:"role"`
			} `graphql:"insert_users_one(object: {username: $userName, email: $email, password: $password, profile: $profile, role:$role, phone:$phone})"`
		}
		mutationVariables := map[string]interface{}{
			"userName": graphql.String(request.Input.UserName),
			"email":    graphql.String(request.Input.Email),
			"password": graphql.String(password),
			"phone":    graphql.String(request.Input.Phone),
			"profile":  graphql.String(profilePictureURL),
			"role":     graphql.String("user"),
		}
		err = client.Mutate(context.Background(), &mutation, mutationVariables)
		if err != nil {
			log.Println("Error inserting user:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
			return
		}
// query newly created user
		var regesteredUserQuery struct {
			User []struct {
				ID       graphql.Int    `graphql:"id"`
				Name     graphql.String `graphql:"username"`
				Email    graphql.String `graphql:"email"`
				Password graphql.String `graphql:"password"`
				Role     graphql.String `graphql:"role"`
				TokenId  graphql.String `graphql:"tokenId"`
			} `graphql:"users(where: {email: {_eq: $email}})"`
		}

		regesteredUserVariable := map[string]interface{}{
			"email": graphql.String(request.Input.Email),
		}
		err = client.Query(context.Background(), &regesteredUserQuery, regesteredUserVariable)
		if err != nil {
			log.Println("Error querying registered user:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query the registered user"})
			return
		}
		if len(regesteredUserQuery.User) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error":"user not found after registration maybe he flead to another table:)"})
		}
		user := regesteredUserQuery.User[0]
// generate email verification token
		emailVerficationToken, err := helpers.GenerateToken()
		
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate a token", "detail": err.Error()})
			return
		}

		var VerficationEmailMutation struct {
			InsertedToken struct {
				ID     graphql.Int    `graphql:"id"`
				Token  graphql.String `graphql:"token"`
			} `graphql:"insert_email_verification_tokens_one(object: {token: $token, user_id: $userId})"`
		}

		// "insert_email_verification_tokens_one(object: {token: $token, user_id: $userId})"
		
		verficatonEmailVariable := map[string]interface{}{
			"token":  graphql.String(emailVerficationToken),
			"userId": graphql.Int(user.ID),
			
		}

		err = client.Mutate(context.Background(), &VerficationEmailMutation, verficatonEmailVariable)
		if err != nil {
			log.Printf("Error inserting email verification token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register verfication token"})
			return
		}


		var UpdateUserTokenMutation struct {
			UpdatedUser struct {
				ID      graphql.Int `graphql:"id"`
				TokenID graphql.Int `graphql:"tokenId"` // Use "tokenId" exactly as it's named in the database
			} `graphql:"update_users_by_pk(pk_columns: {id: $userId}, _set: {tokenId: $tokenId})"`
		}

		updatedUserVariables :=map[string]interface{}{
			"userId": graphql.Int(user.ID),
			"tokenId": VerficationEmailMutation.InsertedToken.ID,
		}

		err = client.Mutate(context.Background(), &UpdateUserTokenMutation, updatedUserVariables)

		if err != nil {
			log.Printf("error updating user with tokenId: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user tokenid", "details": err.Error()})
			return
		}
		log.Printf("User %d successfully updated with tokenId %d", user.ID, UpdateUserTokenMutation.UpdatedUser.TokenID)
	


		// verficationLink := os.Getenv("CHAPA_RETURN_URL") + "?verification token" + emailVerficationToken + "&user_id=" + strconv.Itoa(int(user.ID))
		verificationLink := os.Getenv("CHAPA_RETURN_URL") + "?verification_token=" + emailVerficationToken + "&id=" + strconv.Itoa(int(user.ID))

		// verficationLink := fmt.Sprintf("%s?verification_token=%s&user_id=%d",
		// 	os.Getenv("CHAPA_RETURN_URL"),
		// 	url.QueryEscape(emailVerficationToken),
		// 	user.ID,
		// 	)
		// verficationLink := fmt.Sprintf("http://localhost:5000/api/users/verify-email?verification_token=%s&user_id=%d", emailVerficationToken, user.ID)


		// verficationLink := "?verfication_token" + emailVerficationToken + "&id=" + strconv.Itoa(int(user.ID))

		var emailForm helpers.EmailData
		emailForm.Name = string(user.Name)
		emailForm.Email = string(user.Email)
		emailForm.Link = verificationLink
		emailForm.Subject = "Verifying your email"

		res, errString := helpers.SendEmail(
			[]string{emailForm.Email},
			"verifyEmail.html",
			emailForm,
		)
		if !res {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email verification email", "details": errString})
			return
		}

		token, refreshToken, err := helpers.GenerateAllTokens(string(user.Email), string(user.Name), string(user.Role), string(user.TokenId))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token", "details": err.Error()})
			return

		}

		response := models.SignedUpUserOutput{}
		response.ID = user.ID
		response.Name = user.Name
		response.Email = user.Email
		response.Token = graphql.String(token)
		response.Role = user.Role
		response.RefreshToken = graphql.String(refreshToken)

		c.JSON(http.StatusOK, response)

	}

}


func VerifyEmail() gin.HandlerFunc {
	return func(c *gin.Context){
		client := libs.SetupGraphqlClient()
		var request requests.EmailVerifyRequest

		if err := c.ShouldBindJSON(&request); err !=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":"invalid input", "details": err.Error()})
			return
		}
		// Validate the request body
		validationError := validate.Struct(request)
		if validationError !=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":validationError.Error()})
			return

		}

		// Fetch the token from the database based on the provided 
		var query struct{
			Tokens []struct{
				ID graphql.Int `graphql:"id"`
				Token graphql.String `graphql:"token"`
				UserId graphql.Int `graphql:"user_id"`
				ExpiresAt graphql.String `graphql:"expires_at"`
			}`graphql:"email_verification_tokens(where: {token: {_eq: $token}})"`
		}
		variables := map[string]interface{}{
			"token": graphql.String(request.Input.VerificationToken),
		}
		 err := client.Query(context.Background(), &query, variables)
		 if err !=nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error":"failed to query the verification token", "details": err.Error()})
			return
		 }
// Check if token is found
		 if len(query.Tokens) == 0{
			c.JSON(http.StatusUnauthorized, gin.H{"error":"verification token expired or invalid verification token"})
			return
		 }


		//  validate if the token is expired
		expirationTime, err := time.Parse(time.RFC3339, string(query.Tokens[0].ExpiresAt))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error":"failed to parse token expiration date", "details": err.Error()})
			return
		}

		if time.Now().After(expirationTime){
			c.JSON(http.StatusUnauthorized, gin.H{"error":"verification token expired"})
			return
		}

		 var mutation struct{
			UpdateUser struct{
				ID graphql.ID `graphql:"id"`
				UserName graphql.String `graphql:"username"`
				Email graphql.String `graphql:"email"`
				Profile graphql.String `graphql:"profile"`
				Role graphql.String `graphql:"role"`
			}`graphql:"update_users_by_pk(pk_columns: {id: $id}, _set: {is_email_verified: $status})"`
		 }

		 mutationVariables := map[string]interface{}{
			"id": graphql.Int(request.Input.UserId),
			"status": graphql.Boolean(true),
		 }

		 err = client.Mutate(context.Background(), &mutation, mutationVariables)
		 if err != nil {
			log.Printf("Error updating user email verification status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error":"failed to update user email verification status", "details": err.Error()})
			return
		 }

	
// delete the token after it has been used i will be back for this
		// var deleteMutation struct{
		// 	DeleteToken struct{
		// 		ID graphql.Int `graphql:"id"`
		// 	}`graphql:"delete_email_verification_tokens_by_pk(id: $id)"`
		// }
		// deleteVariables := map[string]interface{}{
		// 	"id":query.Tokens[0].ID,
		// }
		// err = client.Mutate(context.Background(), &deleteMutation, deleteVariables)
		// if err != nil{
		// 	log.Printf("Error deleting email verification token: %v", err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error":"failed to delete email verification token", "details": err.Error()})
		// 	return
		// }

		// return success message

		res := struct {
            Msg string `json:"message"`
        }{
            Msg: "Your email is verified successfully!",
        }

        c.JSON(http.StatusOK, res)
	}
}
