package models

import "github.com/shurcooL/graphql"

type Book struct {
	Title     string      `json:"title"`
	Author    string      `json:"author"`
	Available bool        `json:"available"`
	BookImage *ImageInput `json:"image"`
	Genre     string      `json:"genre"`
}

type CreatedRentBook struct {
	PaymentId  graphql.Int    `json:"payment_id"`
	CheckOutUrl graphql.String `json:"checkout_url"`
	Message    string         `json:"message"`
	UserId     graphql.Int    `json:"user_id"`
	BookId     graphql.Int    `json:"book_id"`
	RentDay    graphql.String `json:"rent_day"`
	Price      graphql.Int    `json:"price"`
	ReturnDate graphql.String `json:"return_date"`
}


type ProcessPaymentOutput struct{
	Message string `json:"message"`
	Status  graphql.String `json:"status"`
}