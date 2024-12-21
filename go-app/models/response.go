package models

import "github.com/shurcooL/graphql"

type SignedUpUserOutput struct {
	ID           graphql.Int    `json:"id"`
	Name         graphql.String `json:"name"`
	Email        graphql.String `json:"email"`
	Token        graphql.String `json:"token"`
	RefreshToken graphql.String `json:"refreshToken"`
	Role         graphql.String `json:"role"`
}
