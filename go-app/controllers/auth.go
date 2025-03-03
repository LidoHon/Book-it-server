package controllers

import (
	"context"
	"fmt"
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
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Declare a variable to hold the input data from the json req body
		var request requests.RegisterRequest

		// Bind the JSON request body to the request struct
		if err := c.ShouldBindJSON(&request); err != nil {
			log.Printf("Validation error: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input data"})
			return
		}

		// Validate the input data
		if err := validate.Struct(request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Validation failed", "details": err.Error()})
			return
		}

		// Check if the user already exists
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

		if err := client.Query(ctx, &query, variable); err != nil {
			log.Println("Error querying existing user:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to query the existing user", "details": err.Error()})
			return
		}

		if len(query.User) != 0 {
			log.Println("User already exists")
			c.JSON(http.StatusBadRequest, gin.H{"message": "User already exists"})
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
		// hash the password
		password := helpers.HashPassword(request.Input.Password)

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
		err := client.Mutate(context.Background(), &mutation, mutationVariables)
		if err != nil {
			log.Println("Failed to register user:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to register user"})
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
			log.Println("Error querying the registered user:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to query the registered user"})
			return
		}
		if len(regesteredUserQuery.User) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "user not found after registration maybe he flead to another table:)"})
		}
		user := regesteredUserQuery.User[0]
		// generate email verification token
		emailVerficationToken, err := helpers.GenerateToken()

		if err != nil {
			log.Println("Error generating email verification token:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to generate email verification token", "detail": err.Error()})
			return
		}

		var VerficationEmailMutation struct {
			InsertedToken struct {
				ID    graphql.Int    `graphql:"id"`
				Token graphql.String `graphql:"token"`
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
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to register verfication token"})
			return
		}

		var UpdateUserTokenMutation struct {
			UpdatedUser struct {
				ID      graphql.Int `graphql:"id"`
				TokenID graphql.Int `graphql:"tokenId"`
			} `graphql:"update_users_by_pk(pk_columns: {id: $userId}, _set: {tokenId: $tokenId})"`
		}

		updatedUserVariables := map[string]interface{}{
			"userId":  graphql.Int(user.ID),
			"tokenId": VerficationEmailMutation.InsertedToken.ID,
		}

		err = client.Mutate(context.Background(), &UpdateUserTokenMutation, updatedUserVariables)

		if err != nil {
			log.Printf("error updating user with tokenId: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update user tokenid", "details": err.Error()})
			return
		}
		log.Printf("User %d successfully updated with tokenId %d", user.ID, UpdateUserTokenMutation.UpdatedUser.TokenID)

		verificationLink := os.Getenv("CHAPA_RETURN_URL") + "?verification_token=" + emailVerficationToken + "&user_id=" + strconv.Itoa(int(user.ID))

		emailForm := helpers.EmailData{
			Name:    string(user.Name),
			Email:   string(user.Email),
			Link:    verificationLink,
			Subject: "Verifying your email",
		}

		res, errString := helpers.SendEmail(
			[]string{emailForm.Email},
			"verifyEmail.html",
			emailForm,
		)
		if !res {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send email verification email, please contact support.", "details": errString})
			return
		}

		token, refreshToken, err := helpers.GenerateAllTokens(string(user.Email), string(user.Name), string(user.Role), string(user.TokenId))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to  generate jwt token", "details": err.Error()})
			return

		}

		response := models.SignedUpUserOutput{
			ID:           user.ID,
			UserName:     user.Name,
			Email:        user.Email,
			Token:        graphql.String(token),
			Role:         user.Role,
			RefreshToken: graphql.String(refreshToken),
		}
		c.JSON(http.StatusOK, response)

	}

}

func VerifyEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()

		var request requests.EmailVerifyRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid input", "details": err.Error()})
			return
		}
		// Validate the request body
		validationError := validate.Struct(request)
		if validationError != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": validationError.Error()})
			return

		}

		// Fetch the token from the database based on the provided
		var query struct {
			Tokens []struct {
				ID        graphql.Int    `graphql:"id"`
				Token     graphql.String `graphql:"token"`
				UserId    graphql.Int    `graphql:"user_id"`
				ExpiresAt graphql.String `graphql:"expires_at"`
			} `graphql:"email_verification_tokens(where: {token: {_eq: $token}})"`
		}
		variables := map[string]interface{}{
			"token": graphql.String(request.Input.VerificationToken),
		}
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to query the verification token", "details": err.Error()})
			return
		}
		// Check if token is found
		if len(query.Tokens) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid verification token"})
			return
		}

		//  validate if the token is expired
		expirationTime, err := time.Parse(time.RFC3339, string(query.Tokens[0].ExpiresAt))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to parse token expiration date", "details": err.Error()})
			return
		}

		if time.Now().After(expirationTime) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "expired verification token"})
			return
		}

		var mutation struct {
			UpdateUser struct {
				ID       graphql.ID     `graphql:"id"`
				UserName graphql.String `graphql:"username"`
				Email    graphql.String `graphql:"email"`
				Profile  graphql.String `graphql:"profile"`
				Role     graphql.String `graphql:"role"`
			} `graphql:"update_users_by_pk(pk_columns: {id: $id}, _set: {is_email_verified: $status})"`
		}

		mutationVariables := map[string]interface{}{
			"id":     graphql.Int(request.Input.UserId),
			"status": graphql.Boolean(true),
		}

		err = client.Mutate(context.Background(), &mutation, mutationVariables)
		if err != nil {
			log.Printf("Error updating user email verification status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update user email verification status", "details": err.Error()})
			return
		}

		// delete the token after it has been used
		var deleteMutation struct {
			DeleteToken struct {
				ID graphql.Int `graphql:"id"`
			} `graphql:"delete_email_verification_tokens_by_pk(id: $id)"`
		}
		deleteVariables := map[string]interface{}{
			"id": query.Tokens[0].ID,
		}
		err = client.Mutate(context.Background(), &deleteMutation, deleteVariables)
		if err != nil {
			log.Printf("Error deleting email verification token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to delete email verification token", "details": err.Error()})
			return
		}

		// return success message

		res := struct {
			Msg string `json:"message"`
		}{
			Msg: "Your email is verified successfully!",
		}

		c.JSON(http.StatusOK, res)
	}
}

// login user

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()

		var request requests.LoginRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
			return
		}

		var query struct {
			User []struct {
				ID              graphql.Int     `graphql:"id"`
				Name            graphql.String  `graphql:"username"`
				Email           graphql.String  `graphql:"email"`
				Password        graphql.String  `graphql:"password"`
				Role            graphql.String  `graphql:"role"`
				TokenId         graphql.Int     `graphql:"tokenId"`
				IsEmailVerified graphql.Boolean `graphql:"is_email_verified"`
			} `graphql:"users(where: {email: {_eq: $email}})"`
		}

		variables := map[string]interface{}{
			"email": graphql.String(request.Input.Email),
		}

		if err := client.Query(context.Background(), &query, variables); err != nil {
			log.Printf("failed to query the user with email %s: %v", request.Input.Email, err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to query user"})
			return
		}

		if len(query.User) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
			return

		}
		user := query.User[0]

		if !user.IsEmailVerified {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "You need to verify your account first to login"})
			return
		}

		if valid, msg := helpers.VerifyPassword(request.Input.Password, string(user.Password)); !valid {
			c.JSON(http.StatusBadRequest, gin.H{"message": msg})
			return
		}

		token, refreshToken, err := helpers.GenerateAllTokens(string(user.Email), string(user.Name), string(user.Role), string(user.TokenId))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to generate tokens",
				"details": err.Error(),
			})
			return
		}

		response := models.LoginResponse{
			User: &models.UserResponse{
				ID:           user.ID,
				Token:        graphql.String(token),
				RefreshToken: graphql.String(refreshToken),
				Email:        graphql.String(user.Email),
				Name:         graphql.String(user.Name),
				Role:         graphql.String(user.Role),
			},
		}

		c.JSON(http.StatusOK, response)

	}
}

func ResetPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()

		var request requests.PasswordResetRequest

		if err := c.ShouldBindJSON(&request); err != nil {
			log.Printf("invalid input:%v", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid input", "details": err.Error()})
			return
		}
		// fetch user by email
		var query struct {
			User []struct {
				ID              graphql.Int     `graphql:"id"`
				Name            graphql.String  `graphql:"username"`
				Email           graphql.String  `graphql:"email"`
				Password        graphql.String  `graphql:"password"`
				Role            graphql.String  `graphql:"role"`
				TokenId         graphql.Int     `graphql:"tokenId"`
				IsEmailVerified graphql.Boolean `graphql:"is_email_verified"`
			} `graphql:"users(where: {email: {_eq: $email}})"`
		}
		queryVars := map[string]interface{}{
			"email": graphql.String(request.Input.Email),
		}

		if err := client.Query(context.Background(), &query, queryVars); err != nil {
			log.Printf("failed to query a user data: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to query user data", "details": err.Error()})
			return
		}

		if len(query.User) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid credentials"})
			return
		}

		user := query.User[0]
		if !user.IsEmailVerified {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Please verify your email first to reset your password"})
			return
		}

		token, err := helpers.GenerateToken()
		if err != nil {
			log.Printf("Failed to generate token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to generate token", "details": err.Error()})
			return
		}
		// Insert reset token into the database
		var mutation struct {
			InsertedToken struct {
				ID    graphql.Int    `graphql:"id"`
				Token graphql.String `graphql:"token"`
			} `graphql:"insert_email_verification_tokens_one(object: {token: $token, user_id: $userId})"`
		}

		mutationVars := map[string]interface{}{
			"token":  graphql.String(token),
			"userId": graphql.Int(user.ID),
		}

		if err := client.Mutate(context.Background(), &mutation, mutationVars); err != nil {
			log.Printf("failed to register reset token")
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to register rewste token", "details": err.Error()})
			return
		}

		// update the user table to add tokenid
		var UpdateUserTokenMutation struct {
			UpdatedUser struct {
				ID      graphql.Int `graphql:"id"`
				TokenID graphql.Int `graphql:"tokenId"`
			} `graphql:"update_users_by_pk(pk_columns: {id: $userId}, _set: {tokenId: $tokenId})"`
		}

		updatedUserVariables := map[string]interface{}{
			"userId":  graphql.Int(user.ID),
			"tokenId": mutation.InsertedToken.ID,
		}

		err = client.Mutate(context.Background(), &UpdateUserTokenMutation, updatedUserVariables)

		if err != nil {
			log.Printf("error updating user with tokenId: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update user tokenid", "details": err.Error()})
			return
		}
		log.Printf("User %d successfully updated with tokenId %d", user.ID, UpdateUserTokenMutation.UpdatedUser.TokenID)

		verificationLink := os.Getenv("RESET_PASS_URL") + "/password-reset?token=" + token + "&id=" + strconv.Itoa(int(user.ID))
		// verificationLink := fmt.Sprintf(
		// 	"%s/password-reset?token=%s&id=%d",
		// 	os.Getenv("RESET_PASS_URL"), // Base URL
		// 	token,                         // Reset token
		// 	int(user.ID),                   // User ID
		// )

		// Send password reset email
		emailData := helpers.EmailData{
			Name:    string(user.Name),
			Email:   string(user.Email),
			Link:    verificationLink,
			Subject: "Reset your password",
		}

		if success, errString := helpers.SendEmail(
			[]string{emailData.Email},
			"passReset.html",
			emailData,
		); !success {
			log.Printf("Failed to send password reset email: %v", errString)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send password reset email", "details": errString})
			return
		}

		response := models.ResetRequestOutput{
			ID:      mutation.InsertedToken.ID,
			Message: "we have sent you an email to reset your password, please go and check your email to reset your password",
		}

		c.JSON(http.StatusOK, response)

	}
}

func UpdatePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()

		var request requests.UpdatePasswordRequest

		if err := c.ShouldBind(&request); err != nil {
			log.Printf("error binding request: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid input", "details": err.Error()})
			return
		}
		log.Printf("Incoming request: %+v", request)
		if validationError := validate.Struct(request); validationError != nil {
			log.Printf("Validation error: %v", validationError)
			c.JSON(http.StatusBadRequest, gin.H{"message": "validation failed", "details": validationError.Error()})
			return
		}

		var query struct {
			Tokens []struct {
				ID     graphql.Int    `graphql:"id"`
				Token  graphql.String `graphql:"token"`
				UserId graphql.Int    `graphql:"user_id"`
			} `graphql:"email_verification_tokens(where: {token: {_eq: $token}})"`
		}
		queryVars := map[string]interface{}{
			"token": graphql.String(request.Input.Token),
		}

		if err := client.Query(context.Background(), &query, queryVars); err != nil {
			log.Printf("failed to query token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to query token", "details": err.Error()})
			return
		}

		if len(query.Tokens) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid tokens"})
			return
		}

		password := helpers.HashPassword(request.Input.Password)

		var mutation struct {
			UpdateUser struct {
				ID       graphql.Int    `graphql:"id"`
				UserName graphql.String `graphql:"username"`
				Email    graphql.String `graphql:"email"`
				Profile  graphql.String `graphql:"profile"`
				Role     graphql.String `graphql:"role"`
			} `graphql:"update_users_by_pk(pk_columns: {id: $id}, _set: {password: $password})"`
		}

		mutationVars := map[string]interface{}{
			"id":       graphql.Int(request.Input.UserId),
			"password": graphql.String(password),
		}

		err := client.Mutate(context.Background(), &mutation, mutationVars)
		if err != nil {
			log.Printf("failed to update the password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update the password", "details": err.Error()})
			return
		}
		emailForm := helpers.EmailData{
			Name:    string(mutation.UpdateUser.UserName),
			Email:   string(mutation.UpdateUser.Email),
			Subject: "Password reseted successfully!",
		}

		sucess, errorString := helpers.SendEmail(
			[]string{emailForm.Email}, "resetPasswordSuccess.html", emailForm,
		)
		if !sucess {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to send password reset email", "details": errorString})
			return
		}

		response := models.UpdatePasswordResponse{
			Message: "your password has been reseted successfully",
		}
		log.Println(response)

		c.JSON(http.StatusOK, response)

		// delete the token after it has been used
		var deleteMutation struct {
			DeleteToken struct {
				ID graphql.Int `graphql:"id"`
			} `graphql:"delete_email_verification_tokens_by_pk(id: $id)"`
		}
		deleteVariables := map[string]interface{}{
			"id": query.Tokens[0].ID,
		}
		err = client.Mutate(context.Background(), &deleteMutation, deleteVariables)
		if err != nil {
			log.Printf("Error deleting email verification token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to delete email verification token", "details": err.Error()})
			return
		}

	}
}

func DeleteUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var eventPayload requests.EventPayload

		if err := c.ShouldBindJSON(&eventPayload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
			return
		}

		if eventPayload.Event.Op == "DELETE" && eventPayload.Event.Data.Old != nil {
			oldUserData := eventPayload.Event.Data.Old

			emailData := helpers.EmailData{
				Name:    oldUserData.UserName,
				Email:   oldUserData.Email,
				Subject: "Account Deletion Confirmation",
			}

			success, errorString := helpers.SendEmail(
				[]string{emailData.Email},
				"deletedAccount.html",
				emailData,
			)

			if !success {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send account deletion email", "details": errorString})
				return
			}

			c.JSON(http.StatusOK, gin.H{"status": "Account removal email has been sent", "email": oldUserData.Email})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"message": "Something went wrong", "details": "Invalid input"})
	}
}

func UpdateProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()

		var req requests.UpdateRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			fmt.Printf("error from jsonbind %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid input", "details": err.Error()})
			return
		}

		if validationErr := validate.Struct(req); validationErr != nil {
			fmt.Println("Validation Error:", validationErr.Error())
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"updateProfile": validationErr.Error(),
			},
			)
			return
		}

		imageUrl, exists := c.Get("imageUrl")
		if !exists {
			imageUrl = ""
		}

		proPicUrl, ok := imageUrl.(string)
		if !ok {
			proPicUrl = ""
		}

		if proPicUrl == "" {
			var mutation struct {
				UpdateProfile struct {
					ID graphql.Int `graphql:"id"`
				} `graphql:"update_users_by_pk(pk_columns: {id: $userId}, _set: {username: $userName, phone: $Phone})"`
			}

			mutationVars := map[string]interface{}{
				"userId":   graphql.Int(req.Input.UserId),
				"userName": graphql.String(req.Input.UserName),
				"Phone":    graphql.String(req.Input.Phone),
			}

			err := client.Mutate(context.Background(), &mutation, mutationVars)
			if err != nil {
				log.Println("Failed to update user profile:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update user profile", "details": err.Error()})
				return
			}

			res := models.UpdateResponce{
				Message: "Profile updated successfully",
			}
			c.JSON(http.StatusOK, res)
		} else {
			var mutation struct {
				UpdateProfile struct {
					ID graphql.Int `graphql:"id"`
				} `graphql:"update_users_by_pk(pk_columns: {id: userId}, _set: {username: $userName, phone: $Phone, profile: $Profile})"`
			}

			mutationVars := map[string]interface{}{
				"userId":   graphql.Int(req.Input.UserId),
				"userName": graphql.String(req.Input.UserName),
				"Phone":    graphql.String(req.Input.Phone),
				"Profile":  graphql.String(proPicUrl),
			}

			err := client.Mutate(context.Background(), &mutation, mutationVars)
			if err != nil {
				log.Println("Failed to update user profile:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update user profile"})
				return
			}
			res := models.UpdateResponce{
				Message: "Profile updated successfully",
			}

			c.JSON(http.StatusOK, res)
		}

	}

}

func DeleteUserByEmail() gin.HandlerFunc {
	return func(c *gin.Context) {

		client := libs.SetupGraphqlClient()

		var input requests.DeleteUserWithEmailInput
		if err := c.ShouldBindBodyWithJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input", "details": err.Error()})
			return
		}

		var query struct {
			User struct {
				ID       graphql.Int    `graphql:"id"`
				Email    graphql.String `graphql:"email"`
				UserName graphql.String `graphql:"username"`
			} `graphql:"users_by_pk(id: $id)"`
		}

		queryVars := map[string]interface{}{
			"id": graphql.Int(input.UserID),
		}

		if err := client.Query(context.Background(), &query, queryVars); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to query user data", "details": err.Error()})
			return
		}

		var mutation struct {
			DeleteUser struct {
				ID graphql.Int `graphql:"id"`
			} `graphql:"delete_users_by_pk(id: $id)"`
		}

		mutationVars := map[string]interface{}{
			"id": graphql.Int(input.UserID),
		}

		if err := client.Mutate(context.Background(), &mutation, mutationVars); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete user", "details": err.Error()})
			return
		}
		emailData := helpers.EmailData{
			Name:    string(query.User.UserName),
			Email:   string(query.User.Email),
			Subject: "Account Deletion Confirmation",
		}

		success, errorString := helpers.SendEmail(
			[]string{emailData.Email},
			"deletedAccount.html", emailData,
		)

		if !success {
			// Undo the deletion if email fails
			var undoMutation struct {
				InsertUser struct {
					ID graphql.Int `graphql:"id"`
				} `graphql:"insert_users_one(object: {id: $id, email: $email, user_name: $userName})"`
			}
			undoVars := map[string]interface{}{
				"id":       graphql.Int(query.User.ID),
				"email":    graphql.String(query.User.Email),
				"userName": graphql.String(query.User.UserName),
			}
			if undoErr := client.Mutate(context.Background(), &undoMutation, undoVars); undoErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to rollback deletion", "details": undoErr.Error()})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to send email, deletion rolled back", "details": errorString})
			return
		}

		c.JSON(http.StatusOK, models.DeleteUserWithEmailResponse{
			Status:  "success",
			Message: "User deleted and email sent successfully",
		})
	}
}
