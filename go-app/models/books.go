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
	UserId  graphql.Int    `json:"user_id"`
	BookId  graphql.Int    `json:"book_id"`
	RentDay graphql.String `json:"rent_day"`
}
