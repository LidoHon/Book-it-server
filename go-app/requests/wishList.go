package requests

import "github.com/shurcooL/graphql"

type CreateWishlistRequest struct {
	ID     graphql.Int `json:"id"`
	UserId graphql.Int `json:"userId"`
	BookId graphql.Int `json:"bookId"`
}
