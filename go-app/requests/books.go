package requests

import (
	"encoding/json"

	"github.com/LidoHon/devConnect/models"
)

type InsertBooksRequest struct {
	Input struct {
		Title     string             `json:"title"`
		Author    string             `json:"author"`
		Available bool               `json:"available"`
		BookImage *models.ImageInput `json:"image"`
		Genre     string             `json:"genre"`
	} `json:"input"`
}

type GetBooksRequest struct {
	Title     string             `json:"title"`
	Author    string             `json:"author"`
	Available bool               `json:"available"`
	BookImage *models.ImageInput `json:"image"`
	Genre     string             `json:"genre"`
}

type UpdateBookRequest struct {
	BookId    json.Number        `json:"book_id"`
	Title     string             `json:"title" validate:"required"`
	Author    string             `json:"author"`
	Genre     string             `json:"genre"`
	BookImage *models.ImageInput `json:"image"`
	Available bool               `json:"available"`
}

type UploadEndpointRequest struct {
	Input struct {
		BookId int `json:"book_id" validate:"required"`
	} `json:"input"`
}
