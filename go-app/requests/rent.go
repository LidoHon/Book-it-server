package requests

// "github.com/shurcooL/graphql"

type PlaceRentRequest struct {
	ID         int    `json:"id"`
	UserId     int    `json:"user_id"`
	BookId     int    `json:"book_id"`
	ReturnDate string `json:"return_date"`
	Price      int    `json:"price"`
}

type UpdateRentRequest struct {
	Input struct {
		ID          int    `json:"id"`
		ReturedDate string `json:"returedDate"`
	} `json:"input"`
}

type PaymentProcessRequest struct {
	Input struct {
		TxRef string `json:"tx_ref"`
		Id    int    `json:"id"`
	} `json:"input"`
}

type ReturnBookRequest struct {
	Input struct {
		BookId int `json:"bookId"`
	} `json:"input"`
}
