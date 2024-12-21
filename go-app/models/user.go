package models

import (
	"time"
)

type User struct {
	ID        *int      `json:"id"`
	UserName  *string   `json:"userName" validate:"required,min=2,max=200"`
	Profile   *string   `json:"profile"`
	Role      *string   `json:"role"`
	Email     *string   `json:"email" validate:"email,required"`
	Password  *string   `json:"password" validate:"required,min=6"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
