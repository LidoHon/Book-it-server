package requests

import (
	"github.com/shurcooL/graphql"
)

type PlaceRentRequest struct {
	ID      graphql.Int `json:"id"`
	UserId  graphql.Int `json:"userId"`
	BookId  graphql.Int `json:"bookId"`
}


type UpdateRentRequest struct{
	ID graphql.Int `json:"id"`
	ReturedDate graphql.String `json:"returedDate"`
}