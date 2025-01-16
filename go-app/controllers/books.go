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

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
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
				ID        int             `json:"id"`
				Title     graphql.String  `json:"title"`
				Author    graphql.String  `json:"author"`
				Available graphql.Boolean `json:"available"`
				BookImage graphql.String  `json:"bookImage"`
				Genre     graphql.String  `json:"genre"`
			} `graphql:" insert_books_one(object: {author: $author, available: $available, bookImage: $image, genre: $genre, title: $title})"`
		}

		mutationVars := map[string]interface{}{
			"title":     graphql.String(req.Input.Title),
			"author":    graphql.String(req.Input.Author),
			"available": graphql.Boolean(req.Input.Available),
			"image":     graphql.String(BookImage),
			"genre":     graphql.String(req.Input.Genre),
		}

		err := client.Mutate(ctx, &mutation, mutationVars)
		if err != nil {
			log.Println("failed to insert a book", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert a book", "details": err.Error()})
			return
		}
		res := models.CreatedBooksOutput{
			ID:        graphql.Int(mutation.InsertBooks.ID),
			Title:     graphql.String(mutation.InsertBooks.Title),
			Author:    graphql.String(mutation.InsertBooks.Author),
			Available: graphql.Boolean(mutation.InsertBooks.Available),
			BookImage: graphql.String(mutation.InsertBooks.BookImage),
			Genre:     graphql.String(mutation.InsertBooks.Genre),
		}
		c.JSON(http.StatusCreated, res)
	}

}

func UpdateBooks() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		bookIdParam := c.Param("id")

		bookId, err := strconv.ParseInt(bookIdParam, 10, 64)
		if err != nil {
			log.Println("failed to convert bookId to int64", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid book ID", "details": err.Error()})
			return
		}
		req := requests.UpdateBookRequest{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
			return
		}

		validationError := validate.Struct(req)
		if validationError != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationError.Error()})
			return
		}

		var query struct {
			Books []struct {
				ID        graphql.Int     `graphql:"id"`
				Title     graphql.String  `graphql:"title"`
				Author    graphql.String  `graphql:"author"`
				Available graphql.Boolean `graphql:"available"`
				BookImage graphql.String  `graphql:"bookImage"`
				Genre     graphql.String  `graphql:"genre"`
			} `graphql:"books(where: {id: {_eq: $id}})"`
		}

		queryVars := map[string]interface{}{
			"id": graphql.Int(bookId),
		}

		err = client.Query(ctx, &query, queryVars)
		if err != nil {
			log.Println("failed to query a book", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to query a book", "details": err.Error()})
			return
		}

		if len(query.Books) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": "book not found"})
			return

		}

		book := query.Books[0]

		imageUrl, exists := c.Get("imageUrl")
		if !exists {
			imageUrl = ""
		}

		BookImage, ok := imageUrl.(string)
		if !ok {
			BookImage = string(book.BookImage)
		}

		var mutation struct {
			UpdateBooks struct {
				ID        int             `json:"id"`
				Title     graphql.String  `json:"title"`
				Author    graphql.String  `json:"author"`
				Available graphql.Boolean `json:"available"`
				BookImage graphql.String  `json:"image"`
				Genre     graphql.String  `json:"genre"`
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
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update a book", "details": err.Error()})
			return
		}

		res := models.UpdatedBookOutput{
			ID:        graphql.Int(mutation.UpdateBooks.ID),
			Title:     mutation.UpdateBooks.Title,
			Author:    mutation.UpdateBooks.Author,
			Available: mutation.UpdateBooks.Available,
			BookImage: mutation.UpdateBooks.BookImage,
			Genre:     mutation.UpdateBooks.Genre,
		}

		c.JSON(200, res)
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
			"limit":  graphql.Int(limit),
			"offset": graphql.Int(offset),
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
		client := libs.SetupGraphqlClient()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()
		bookIdParam := c.Param("id")
		bookId, err := strconv.ParseInt(bookIdParam, 10, 64)

		if err != nil {
			log.Println("failed to convert bookId to int64", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid book id", "details": err.Error()})
			return
		}

		query := struct {
			Book struct {
				ID        graphql.Int     `graphql:"id"`
				Title     graphql.String  `graphql:"title"`
				Author    graphql.String  `graphql:"author"`
				Available graphql.Boolean `graphql:"available"`
				BookImage graphql.String  `graphql:"bookImage"`
				Genre     graphql.String  `graphql:"genre"`
			} `graphql:"books_by_pk(id: $id)"`
		}{}
		queryVars := map[string]interface{}{
			"id": graphql.Int(bookId),
		}

		err = client.Query(ctx, &query, queryVars)
		if err != nil {
			log.Println(" error querying book by id", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "error qurying the book by a given id ", "details": err.Error()})
			return
		}

		if query.Book.ID == 0 {
			log.Println("book not found", err)
			c.JSON(http.StatusNotFound, gin.H{"message": "book not found"})
			return
		}

		c.JSON(200, gin.H{
			"message": "book fetched sucessfully",
			"book": gin.H{
				"id":        query.Book.ID,
				"title":     query.Book.Title,
				"author":    query.Book.Author,
				"available": query.Book.Available,
				"bookImage": query.Book.BookImage,
				"genre":     query.Book.Genre,
			},
		})
	}
}

func DeleteBooks() gin.HandlerFunc {
	return func(c *gin.Context) {

		client := libs.SetupGraphqlClient()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		bookIdParam := c.Param("id")
		bookId, err := strconv.ParseInt(bookIdParam, 10, 64)
		if err != nil {
			log.Println("failed to convert book id to int64", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid book id", "details": err.Error()})
			return
		}

		query := struct {
			Book struct {
				ID        graphql.Int     `graphql:"id"`
				Title     graphql.String  `graphql:"title"`
				Author    graphql.String  `graphql:"author"`
				Available graphql.Boolean `graphql:"available"`
				BookImage graphql.String  `graphql:"bookImage"`
				Genre     graphql.String  `graphql:"genre"`
			} `graphql:"books_by_pk(id: $id)"`
		}{}
		queryVars := map[string]interface{}{
			"id": graphql.Int(bookId),
		}

		err = client.Query(ctx, &query, queryVars)
		if err != nil {
			log.Println("error querying the book by id", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "error querying the book by id", "details": err.Error()})
			return
		}
		if query.Book.ID == 0 {
			log.Println("book not found")
			c.JSON(http.StatusNotFound, gin.H{"message": "book not found"})
			return
		}

		var mutation struct {
			DeletedBook struct {
				ID graphql.Int `graphql:"id"`
			} `graphql:"delete_books_by_pk(id: $id)"`
		}

		mutateVars := map[string]interface{}{
			"id": graphql.Int(bookId),
		}

		err = client.Mutate(ctx, &mutation, mutateVars)
		if err != nil {
			log.Println("error deleting the book by id", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "error deleting the book by id",
				"details": err.Error()})
			return
		}
		if mutation.DeletedBook.ID == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "book not found",
			})
		}

		c.JSON(200, gin.H{
			"message": "book deleted successfully",
			"bookId":  mutation.DeletedBook.ID,
		})
	}
}
