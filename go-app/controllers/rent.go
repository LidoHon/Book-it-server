package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/LidoHon/devConnect/libs"
	"github.com/LidoHon/devConnect/models"
	"github.com/LidoHon/devConnect/requests"
	"github.com/gin-gonic/gin"
	"github.com/shurcooL/graphql"
)

var mu sync.Mutex

func PlaceRent() gin.HandlerFunc {
	return func(c *gin.Context) {

		client := libs.SetupGraphqlClient()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		req := requests.PlaceRentRequest{}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid input", "details": err.Error()})
			return
		}

		validationError := validate.Struct(req)
		if validationError != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": validationError.Error()})
			return
		}

		mu.Lock() 
		defer mu.Unlock()

		// Check if the book is already rented (pseudo-code for locking logic)
		var query struct {
			Book struct {
				Available bool `json:"available"`
			} `graphql:"books_by_pk(id: $id)"`
		}

		queryVars := map[string]interface{}{
			"id": graphql.Int(req.BookId),
		}

		err := client.Query(ctx, &query, queryVars)
		if err != nil {
			log.Println("failed to check book availability", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to check book availability", "details": err.Error()})
			return
		}

		if !query.Book.Available {
			c.JSON(http.StatusConflict, gin.H{"message": "book is already rented"})
			return
		}



		var mutation struct {
			RentBook struct {
				UserId  int       `json:"userId"`
				BookId  int       `json:"bookId"`
			} `graphql:"insert_rentedBooks_one(object: {bookId: $bookId, userId: $userId}) "`
		}
		mutationVars := map[string]interface{}{
			"bookId":  graphql.Int(req.BookId),
			"userId":  graphql.Int(req.UserId),
			 
		}

		err = client.Mutate(ctx, &mutation, mutationVars)
		if err != nil {
			log.Println("faiked to place a rent", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create a rent", "details": err.Error()})
			return
		}

		mutationBooks := struct{
			UpdateBookByPk struct {
				Available bool `json:"available"`
			}`graphql:"update_books_by_pk(pk_columns: {id: $id}, _set: {available: false})"`
		}{}
		

		mutaionBooksVars := map[string]interface{}{
			"id": graphql.Int(req.BookId),
		}

		err = client.Mutate(ctx, &mutationBooks, mutaionBooksVars)
		if err != nil {
			log.Println("faiked to place a rent", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create a rent", "details": err.Error()})
			return
		}

		res := models.CreatedRentBook{
			UserId:  graphql.Int(mutation.RentBook.UserId),
			BookId:  graphql.Int(mutation.RentBook.BookId),
			
		}

		c.JSON(http.StatusOK, res)
	}
}



func DeleteRent () gin.HandlerFunc{
	return func(c *gin.Context){
		client := libs.SetupGraphqlClient()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		rentIdParam := c.Param("id")
		rentId, err := strconv.ParseInt(rentIdParam, 10, 64)
		if err != nil {
			log.Println("failed to convert book id to int64", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid book id", "details": err.Error()})
			return
		}

		query := struct{
			Rent struct{
				ID graphql.Int `graphql:"id"`
				BookId graphql.Int `graphql:"bookId"`
				UserId graphql.Int `graphql:"userId"`
			}`graphql:" rentedBooks_by_pk(id: $id)"`
		}{}
		
		queryVars := map[string]interface{}{
			"id": graphql.Int(rentId),
		}

		err = client.Query(ctx, &query, queryVars)
		if err != nil {
			log.Println("failed to delete a rent", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to delete a rent", "details": err.Error()})
			return
		}

		if query.Rent.ID == 0{
			log.Println("rent not found")
			c.JSON(http.StatusNotFound, gin.H{"message": "rent not found"})
			return
		}

		mutation := struct {
			DeleteRent struct{
				ID graphql.Int `json:"id"`
			}`graphql:" delete_rentedBooks_by_pk(id: $id)"`
		}{}

		mutateVars := map[string]interface{}{
			"id": graphql.Int(rentId),
		}

		if err := client.Mutate(ctx, &mutation, mutateVars); err != nil {
			log.Println("unable to delete a rent", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to delete a rent", "details": err.Error()})
			return
		}

		if mutation.DeleteRent.ID == 0{
			log.Println("rent not found")
			c.JSON(http.StatusNotFound, gin.H{"message": "rent not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "rent deleted sucessfully"})
	}
	
}


func ReturnBook() gin.HandlerFunc {
	return func(c *gin.Context) {

		client := libs.SetupGraphqlClient()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var req struct {
            RentId  int `json:"rentId"`
            BookId  int `json:"bookId"`
        }
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid input", "details": err.Error()})
		}
		// returnDate := time.Now().Format(time.RFC3339)
		// mutation := struct {
        //     UpdateRentedBook struct {
        //         Id int `json:"id"`
        //     } `graphql:"update_rentedBooks_by_pk(pk_columns: {id: $rentId}, _set: {return_date: $returnDate})"` 
        // }{}
        // mutateVars := map[string]interface{}{
        //     "rentId": graphql.Int(req.RentId),
		// 	"returnDate": returnDate,
            
        // }


		// if err := client.Mutate(ctx, &mutation, mutateVars); err !=nil{
		// 	log.Println("failed to return a book", err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to return a book", "details": err.Error()})
		// 	return
		// }

		mutate := struct{
            UpdateBook struct {
                Id int `json:"id"`
            } `graphql:"update_books_by_pk(pk_columns: {id: $bookId}, _set: {available: true})"` 
		}{}
		mutationVars := map[string]interface{}{
			"bookId": graphql.Int(req.BookId),
		}
		err := client.Mutate(ctx, &mutate, mutationVars)
		if err != nil {
			log.Println("failed to return a book", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to return a book", "details": err.Error()})
			return
		}

		// Fetch the book title and check if the book is in any user's wishlist

		var wishlistQuery struct {
			UsersWithBookWishlist []struct {
				UserId int `json:"userId"`
			}`graphql:"users(where: {wishlist: {bookId: {_eq: $bookId}}})"`
			Book struct{
				Title string `json:"title"`
			}`graphql:"books_by_pk(id: $bookId)"`
		}

		wishlistVars := map[string]interface{}{
			"bookId": graphql.Int(req.BookId),
		}

		if err := client.Query(ctx, &wishlistQuery, wishlistVars); err != nil {
			log.Println("failed to fetch book details", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch book details", "details": err.Error()})
			return
		}

		// Send notifications to users who have the book in their wishlist

		for _, user := range wishlistQuery.UsersWithBookWishlist {
			notificationMutation := struct {
				InsertNotification struct {
					Id int `json:"id"`
					
				}`graphql:" insert_notification_one(object: {id: 10, message: $message, userId: $userId})"`
			}{}
			notificationVars := map[string]interface{}{
				"userId": graphql.Int(user.UserId),
				"message": fmt.Sprintf("Book '%s' is now available for rent", wishlistQuery.Book.Title),
			}

			if err := client.Mutate(ctx, &notificationMutation, notificationVars); err != nil {
				log.Println("failed to send notification", err)
				// c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to send notification", "details": err.Error()})
				// return
				 // Continue with other users even if one fails
				continue
		}

		
		
		

		c.JSON(http.StatusOK, gin.H{"message": "book rented and notification sent successfully"})
	}
}


}