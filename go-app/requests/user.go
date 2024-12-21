package requests

import (
	"github.com/LidoHon/devConnect/models"
)

type RegisterRequest struct {
	Input struct {
		UserName string             `json:"userName" validate:"required"`
		Email    string             `json:"email" validate:"required,email"`
		Password string             `json:"password" validate:"required,min=6"`
		Phone    string             `json:"phone" validate:"required"`
		Image    *models.ImageInput `json:"image"`
	} `json:"input"`
}

type EmailVerifyRequest struct{
	Input struct{
		VerificationToken string `json:"verification_token" validate:"required"`
		UserId  int `json:"user_id" validate:"required"`
	}`json:"input"`
}
