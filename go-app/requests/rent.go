package requests

import (
	"github.com/shurcooL/graphql"
)

type PlaceRentRequest struct {
	ID         int    `json:"id"`
	UserId     int    `json:"user_id"`
	BookId     int    `json:"book_id"`
	ReturnDate string `json:"return_date"`
	Price      int    `json:"price"`
}

type UpdateRentRequest struct {
	ID          graphql.Int    `json:"id"`
	ReturedDate graphql.String `json:"returedDate"`
}

type PaymentProcessRequest struct {
	Input struct{
		TxRef string `json:"tx_ref"`
		Id    int    `json:"id"`
	}`json:"input"`
}