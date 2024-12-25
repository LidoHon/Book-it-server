package requests

import (
	"time"

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

type EmailVerifyRequest struct {
	Input struct {
		VerificationToken string `json:"verification_token" validate:"required"`
		UserId            int    `json:"user_id" validate:"required"`
	} `json:"input"`
}

type LoginRequest struct {
	Input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	} `json:"input"`
}

type PasswordResetRequest struct {
	Input struct {
		Email string `json:"email" validate:"required,email"`
	} `json:"input"`
}

type UpdatePasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	UserId   int    `json:"user_id" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

type EventPayload struct {
	CreatedAt    time.Time    `json:"created_at"`
	DeliveryInfo DeliveryInfo `json:"delivery_info"`
	Event        Event        `json:"event"`
	ID           string       `json:"id"`
	Table        Table        `json:"table"`
	Trigger      Trigger      `json:"trigger"`
}

type Trigger struct {
	Name string `json:"name"`
}

type Table struct {
	Name   string `json:"name"`
	Schema string `json:"schema"`
}

type DeliveryInfo struct {
	CurrentRetry int `json:"current_retry"`
	MaxRetries   int `json:"max_retries"`
}

type Event struct {
	Data         EventData    `json:"data"`
	Op           string       `json:"op"`
	TraceContext TraceContext `json:"trace_context"`
}

type EventData struct {
	New *UserData `json:"new"`
	Old *UserData `json:"old"`
}

type UserData struct {
	CreatedAt time.Time `json:"created_at"`
	Email     string    `json:"email"`
	ID        int       `json:"id"`
	Password  string    `json:"password"`
	Phone     string    `json:"phone"`
	Profile   string    `json:"profile"`
	Role      string    `json:"role"`
	TokenId   string    `json:"tokenId"`
	UpdatedAt time.Time `json:"updated_at"`
	UserName  string    `json:"user_name"`
}
type TraceContext struct {
	TraceID string `json:"trace_id"`
	SpanID  string
}

type UpdateRequest struct {
	UserName string             `json:"userName" validate:"required"`
	Phone    string             `json:"phone" validate:"required"`
	UserId   int                `json:"user_id" validate:"required"`
	Image    *models.ImageInput `json:"image"`
}

type DeleteUserRequest struct {
	Input struct {
		ID int `json:"id" validate:"required"`
	}
}

type DeleteUserWithEmailInput struct {
	UserID int `json:"userId"`
}
