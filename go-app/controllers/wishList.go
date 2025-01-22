package controllers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/LidoHon/devConnect/libs"
	"github.com/LidoHon/devConnect/requests"
	"github.com/gin-gonic/gin"
	"github.com/shurcooL/graphql"
)


type HasuraWishlistActionRequest struct {
	Input requests.CreateWishlistRequest `json:"input"`
}
func CreateWishlist() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var WishlistReq HasuraWishlistActionRequest
		if err := c.ShouldBindJSON(&WishlistReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid input", "details": err.Error()})
			return
		}
req := WishlistReq.Input
		validationError := validate.Struct(req)
		if validationError != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": validationError.Error()})
			return
		}

		var mutation struct {
			InsertWishlist struct {
				ID     int `json:"id"`
				UserId int `json:"userId"`
				BookId int `json:"bookId"`
			} `graphql:"insert_wishlist_one(object: {userId: $user_id, bookId: $book_id})"`
		}

		mutationVars := map[string]interface{}{
			"user_id": graphql.Int(req.UserId),
			"book_id": graphql.Int(req.BookId),
		}
		log.Println("GraphQL Mutation Query:", mutation)
		log.Println("Mutation Variables:", mutationVars)

		if err := client.Mutate(ctx, &mutation, mutationVars); err != nil {
			log.Println("failed to create wishlist", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create wishlist", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Wishlist created",
			"data":    mutation.InsertWishlist,
		})
	}
}

// Define a handler for adding a book to the wishlist

func UpdateWishlist() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.JSON(http.StatusOK, gin.H{"message": "Wishlist updated"})
	}
}

func GetWishlistByUserId() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Wishlist"})
	}
}

func GetWishlistById() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Wishlist by id"})
	}
}

func DeleteWishList() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		wishListParam := c.Param("id")
		wishListId, err := strconv.ParseInt(wishListParam, 10, 64)
		if err != nil {
			log.Println("failed to convert book id to int64", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid book id", "details": err.Error()})
			return
		}

		query := struct {
			Wishlist struct {
				ID     graphql.Int `graphql:"id"`
				BookId graphql.Int `graphql:"bookId"`
				UserId graphql.Int `graphql:"userId"`
			} `graphql:" wishlist_by_pk(id: $id)"`
		}{}

		queryVars := map[string]interface{}{
			"id": graphql.Int(wishListId),
		}

		err = client.Query(ctx, &query, queryVars)
		if err != nil {
			log.Println("failed to delete a wishlist", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to delete wishlist", "details": err.Error()})
			return
		}

		if query.Wishlist.ID == 0 {
			log.Println("wish list not found", err)
			c.JSON(http.StatusNotFound, gin.H{"message": "wishlist not found"})
			return
		}

		mutation := struct {
			DeleteWishList struct {
				ID int `json:"id"`
			} `graphql:"delete_wishlist_by_pk(id: $id)"`
		}{}

		mutationVars := map[string]interface{}{
			"id": graphql.Int(wishListId),
		}

		if err := client.Mutate(ctx, &mutation, mutationVars); err != nil {
			log.Println("failed to delete a wishlist", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to delete wishlist", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Wishlist deleted sucessfully"})
	}
}
