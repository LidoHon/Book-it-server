package requests

import (
	"github.com/shurcooL/graphql"
)

type PlaceRentRequest struct {
	UserId int `json:"user_id"`
	BookId int `json:"book_id"`
}

type UpdateRentRequest struct {
	ID          graphql.Int    `json:"id"`
	ReturedDate graphql.String `json:"returedDate"`
}
