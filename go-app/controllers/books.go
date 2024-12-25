package controllers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/LidoHon/devConnect/libs"
	"github.com/LidoHon/devConnect/models"
	"github.com/LidoHon/devConnect/requests"
	"github.com/gin-gonic/gin"
	"github.com/shurcooL/graphql"
)

func InsertBooks() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()

		_, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		req := requests.InsertBooksRequest{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
			return
		}

		validationErr := validate.Struct(req)
		if validationErr != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": validationErr.Error(),
			},
			)
			return
		}

		imageUrl, exists := c.Get("imageUrl")
		if !exists {
			imageUrl = ""
		}

		BookImage, ok := imageUrl.(string)
		if !ok {
			BookImage = ""
		}

		var mutation struct {
			InsertBooks struct {
				ID             int             `json:"id"`
				Title          graphql.String  `json:"title"`
				Author         graphql.String  `json:"author"`
				Available      graphql.Boolean `json:"available"`
				BookImage      graphql.String  `json:"image"`
				Genre graphql.String `json:"genre"`
			} `graphql:" insert_books_one(object: {author: $author, available: $available, bookImage: $image, genre: $genre, title: $title})"`
		}

		mutationVars := map[string]interface{}{
			"title":     graphql.String(req.Title),
			"author":    graphql.String(req.Author),
			"available": graphql.Boolean(req.Available),
			"image":     graphql.String(BookImage),
			"genre":     graphql.String(req.Genre),
		}

		err := client.Mutate(context.Background(), &mutation, mutationVars)
		if err != nil {
			log.Println("failed to insert a book", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert a book", "details": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"message": "books inserted successfully",
			"book":    mutation.InsertBooks,
		})
	}

}

func UpdateBooks() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		req:= requests.UpdateBookRequest{}

		 if err := c.ShouldBindJSON(&req); err != nil {
			 c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details":err.Error()})
			 return
		 }

		 validationError := validate.Struct(req)
		 if validationError != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationError.Error()})
			return
		 }

		 var query struct{
				Books []struct{
					ID graphql.Int `graphql:"id"`
					Title graphql.String `graphql:"title"`
					Author graphql.String `graphql:"author"`
					Available graphql.Boolean `graphql:"available"`
					BookImage graphql.String `graphql:"bookImage"`
					Genre graphql.String `graphql:"genre"`
				}`graphql:"books(where: {id: {_eq: $id}})"`
		 }

		 bookId, err := req.BookId.Int64()
		 if err != nil {
			log.Println("failed to convert bookId to int64", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to convert bookId to int64", "details": err.Error()})
		 }

		 queryVars := map[string]interface{}{
			"bookId": graphql.Int(bookId),
		 }

		 err = client.Query(ctx, &query, queryVars)
		 if err != nil {
			log.Println("failed to query a book", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query a book", "details": err.Error()})
			return
		 }

		 if len(query.Books) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
			return

		 }

		 book := query.Books[0]
		 imageUrl, exists := c.Get("imageUrl")
		 if !exists{
			imageUrl = ""
		 }

		 BookImage, ok := imageUrl.(string)
		 if !ok {
			BookImage = string(book.BookImage)
		 }

		 var mutation struct{
				UpdateBooks struct{
					ID             int             `json:"id"`
					Title          graphql.String  `json:"title"`
					Author         graphql.String  `json:"author"`
					Available      graphql.Boolean `json:"available"`
					BookImage      graphql.String  `json:"image"`
					graphql.String `json:"genre"`
					Genre graphql.String `json:"genre"`
				} `graphql:"update_books_by_pk(pk_columns: {id: $id}, _set: {author: $author, available: $available, bookImage: $image, genre: $genre, title: $title})"`
		 }
		 mutationVars := map[string]interface{}{
			"id":        graphql.Int(book.ID),
			"title":     graphql.String(req.Title),
			"author":    graphql.String(req.Author),
			"available": graphql.Boolean(req.Available),
			"image":     graphql.String(BookImage),
			"genre":     graphql.String(req.Genre),
		 }
		 err = client.Mutate(ctx, &mutation, mutationVars)
		 if err != nil {
			log.Println("failed to update a book", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update a book", "details": err.Error()})
			return
		 }

		 res := models.CreatedBooksOutput{
			ID:       graphql.Int(mutation.UpdateBooks.ID) ,
			Title:     mutation.UpdateBooks.Title,
			Author:    mutation.UpdateBooks.Author,
			Available: mutation.UpdateBooks.Available,
			BookImage: mutation.UpdateBooks.BookImage,
			Genre:     mutation.UpdateBooks.Genre,
		 }
		 

		c.JSON(200,res)
	}
}

func GetBooks() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()

		limitStr := c.DefaultQuery("limit", "10")    
		offsetStr := c.DefaultQuery("offset", "0")   

		limit, err := strconv.Atoi(limitStr)
		if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value"})
				return
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset value"})
				return
		}

		query := struct {
			Books []struct {
				ID        graphql.Int     `graphql:"id"`
				Title     graphql.String  `graphql:"title"`
				Author    graphql.String  `graphql:"author"`
				Available graphql.Boolean `graphql:"available"`
				BookImage graphql.String  `graphql:"bookImage"`
				Genre     graphql.String  `graphql:"genre"`
			} `graphql:"books(limit: $limit, offset: $offset)"`
		}{}

		queryVars := map[string]interface{}{
			"limit":       graphql.Int(limit),
			"offset":      graphql.Int(offset),
		}


		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := client.Query(ctx, &query, queryVars); err != nil {
			log.Println("failed to query books", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query books", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "books fetched successfully",
			"books":   query.Books,
		})
	}
}

func GetBooksById() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.JSON(200, gin.H{"message": "GetBooksById"})
	}
}

func DeleteBooks() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "DeleteBooks"})
	}
}
