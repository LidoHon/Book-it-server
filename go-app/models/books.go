package models

type Book struct {
	Title     string      `json:"title"`
	Author    string      `json:"author"`
	Available bool        `json:"available"`
	BookImage *ImageInput `json:"image"`
	Genre     string      `json:"genre"`
}
