package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/LidoHon/devConnect/helpers"
	"github.com/LidoHon/devConnect/libs"
	"github.com/LidoHon/devConnect/models"
	"github.com/LidoHon/devConnect/requests"
	"github.com/gin-gonic/gin"
	"github.com/shurcooL/graphql"
)

var mu sync.Mutex

type HasuraActionRequest struct {
	Input requests.PlaceRentRequest `json:"input"`
}

func PlaceRent() gin.HandlerFunc {
    return func(c *gin.Context) {
        client := libs.SetupGraphqlClient()

        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()

        var hasuraReq HasuraActionRequest

        if err := c.ShouldBindJSON(&hasuraReq); err != nil {
            log.Println("Failed to bind JSON:", err)
            c.JSON(http.StatusBadRequest, gin.H{"message": "invalid input", "details": err.Error()})
            return
        }
        req := hasuraReq.Input
        returnDate, err := time.Parse("2006-01-02", req.ReturnDate)
        if err != nil {
            log.Println("Failed to parse return date:", err)
            c.JSON(http.StatusBadRequest, gin.H{"message": "invalid return date", "details": err.Error()})
            return
        }

        rentDate := time.Now()
        days := int(returnDate.Sub(rentDate).Hours() / 24)
        if days <= 0 {
            log.Println("Invalid return date, must be after a rent day")
            c.JSON(http.StatusBadRequest, gin.H{"message": "invalid return date, must be after a rent day"})
            return
        }

        // calculate price
        price := int(days) * 1
        log.Println("calculated price:", price)
        log.Println("Request Data from json binding:", req)

        validationError := validate.Struct(req)
        if validationError != nil {
            log.Println("Validation failed:", validationError.Error())
            c.JSON(http.StatusBadRequest, gin.H{"message": validationError.Error()})
            return
        }

        mu.Lock()
        defer mu.Unlock()
        // query user detail
        var userQuery struct {
            User struct {
                Name  string `graphql:"username"`
                Phone string `graphql:"phone"`
                Email string `graphql:"email"`
            } `graphql:"users_by_pk(id: $id)"`
        }
        userQueryVars := map[string]interface{}{
            "id": graphql.Int(req.UserId),
        }

        err = client.Query(ctx, &userQuery, userQueryVars)
        if err != nil {
            log.Println("Failed to fetch user details:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to fetch user details", "details": err.Error()})
            return
        }

        // Check if the book is already rented
        var query struct {
            Book struct {
                Available graphql.Boolean `graphql:"available"`
            } `graphql:"books_by_pk(id: $id)"`
        }

        queryVars := map[string]interface{}{
            "id": graphql.Int(req.BookId),
        }

        log.Println("Checking book availability for Book ID:", req.BookId)

        err = client.Query(ctx, &query, queryVars)
        if err != nil {
            log.Println("Failed to check book availability:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to check book availability", "details": err.Error()})
            return
        }

        log.Println("Book availability status:", query.Book.Available)

        if !query.Book.Available {
            log.Println("Book is already rented.")
            c.JSON(http.StatusConflict, gin.H{"message": "book is already rented"})
            return
        }

        // Proceed with the mutation to rent the book
        var mutation struct {
            RentBook struct {
                ID         graphql.Int    `json:"id"`
                UserId     graphql.Int    `graphql:"userId"`
                BookId     graphql.Int    `graphql:"bookId"`
                ReturnDate graphql.String `graphql:"return_date"`
                Price      graphql.Int    `graphql:"price"`
            } `graphql:"insert_rentedBooks_one(object: {bookId: $bookId, userId: $userId, return_date: $returnDate, price: $price}) "`
        }

        mutationVars := map[string]interface{}{
            "bookId": graphql.Int(req.BookId),
            "userId": graphql.Int(req.UserId),
            "returnDate": graphql.String(returnDate.Format(time.RFC3339)),
            "price":      graphql.Int(price),
        }

        log.Println("Attempting to rent book:", req.BookId, "for user:", req.UserId)

        err = client.Mutate(ctx, &mutation, mutationVars)
        if err != nil {
            log.Println("Failed to place a rent:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create a rent", "details": err.Error()})
            return
        }

        var chapaForm helpers.PaymentRequest
        chapaForm.Amount = price
        chapaForm.Email = userQuery.User.Email
        chapaForm.Currency = "ETB"
        chapaForm.FirstName = userQuery.User.Name
        chapaForm.PhoneNumber = userQuery.User.Phone
        chapaForm.TxRef = fmt.Sprintf("rent-%d", mutation.RentBook.ID)
        chapaForm.ReturnURL = os.Getenv("CHAPA_REDIRECT_URL")
        chapaForm.CallbackURL = os.Getenv("CHAPA_CALLBACK_URL")
        chapaForm.CustomizationTitle = "Book Rent"
        chapaForm.CustomizationDesc = "Payment for book rent"
        log.Println("Book rented successfully:", mutation.RentBook)

        paymentRespose, err := helpers.InitPayment(&chapaForm)
        fmt.Println("chapa form to get the user details......", chapaForm)
        if err != nil {
            log.Println("Failed to initiate payment:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to initiate payment", "details": err.Error()})
            return
        }
        log.Println("Payment initiated successfully:", paymentRespose)
		log.Println("ChapaResponseeeee: ", paymentRespose.ChapaResponse)

        // Insert the payment record
		var paymentID graphql.Int
		var checkoutURL graphql.String
		if paymentRespose.Status {
			// Access checkout_url from the ChapaResponse map
			data, ok := paymentRespose.ChapaResponse["data"].(map[string]interface{})
			if !ok {
				log.Println("Failed to retrieve checkout_url from ChapaResponse")
				c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to retrieve checkout_url from response"})
				return
			}
			checkoutURL, ok := data["checkout_url"].(string)
			if !ok {
				log.Println("Checkout URL not found in data:", data)
				c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to retrieve checkout_url from data"})
				return
			}
		
			log.Println("Checkout URL:", checkoutURL)
		
			// Proceed with storing payment information
			paymentMutation := struct {
				InsertPayment struct {
					ID            graphql.Int    `json:"id"`
					RentId        graphql.Int    `graphql:"rent_id"`
					TxRef         string         `graphql:"tx_ref"`
					CheckoutURL   string         `graphql:"checkout_url"`
					Amount        graphql.Int    `graphql:"amount"`
					Currency      string         `graphql:"currency"`
					PaymentMethod string         `graphql:"payment_method"`
					Status        string         `graphql:"status"`
					UserId        graphql.Int    `graphql:"user_id"`
				} `graphql:"insert_payments_one(object: {rent_id: $rentId, tx_ref: $txRef, checkout_url: $checkoutUrl, amount: $amount, currency: $currency, payment_method: $paymentMethod, status: $status, user_id: $userId})"`
			}{}
		
			paymentVars := map[string]interface{}{
				"rentId":      graphql.Int(mutation.RentBook.ID),
				"txRef":       graphql.String(chapaForm.TxRef),
				"checkoutUrl": graphql.String(checkoutURL), 
				"amount":      graphql.Int(price),
				"currency":    graphql.String(chapaForm.Currency),
				"paymentMethod": graphql.String("card"), 
				"status":      graphql.String("pending"), 
				"userId":      graphql.Int(req.UserId),
			}
		
			err = client.Mutate(ctx, &paymentMutation, paymentVars)
			if err != nil {
				log.Println("Failed to insert payment record:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to insert payment record", "details": err.Error()})
				return
			}
		
			log.Println("Payment record inserted successfully with id:." , paymentMutation.InsertPayment.ID)
			paymentID = paymentMutation.InsertPayment.ID
			checkoutURL = paymentMutation.InsertPayment.CheckoutURL
			log.Println("Payment record inserted successfully and checkout url:." , checkoutURL)

		} else {
			log.Println("Payment initiation failed:", paymentRespose.Message)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "payment initiation failed", "details": paymentRespose.Message})
		}
        // Update book availability to false
        mutationBooks := struct {
            UpdateBookByPk struct {
                Available bool `json:"available"`
            } `graphql:"update_books_by_pk(pk_columns: {id: $id}, _set: {available: false})"`
        }{}

        mutaionBooksVars := map[string]interface{}{
            "id": graphql.Int(req.BookId),
        }

        log.Println("Updating book availability to false for Book ID:", req.BookId)
		

        err = client.Mutate(ctx, &mutationBooks, mutaionBooksVars)
        if err != nil {
            log.Println("Failed to update book availability:", err)
            c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update book availability", "details": err.Error()})
            return
        }

        log.Println("Book availability updated successfully.")

        res := models.CreatedRentBook{
            Message:    "book rented successfully",
			PaymentId:  graphql.Int(paymentID),
            Price:      graphql.Int(price),
            UserId:     graphql.Int(mutation.RentBook.UserId),
            BookId:     graphql.Int(mutation.RentBook.BookId),
            ReturnDate: graphql.String(returnDate.Format(time.RFC3339)),
			CheckOutUrl: graphql.String(checkoutURL),
        }

        log.Println("Returning response:", res)

        c.JSON(http.StatusOK, res)
    }
}


func DeleteRent() gin.HandlerFunc {
	return func(c *gin.Context) {
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

		query := struct {
			Rent struct {
				ID     graphql.Int `graphql:"id"`
				BookId graphql.Int `graphql:"bookId"`
				UserId graphql.Int `graphql:"userId"`
			} `graphql:" rentedBooks_by_pk(id: $id)"`
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

		if query.Rent.ID == 0 {
			log.Println("rent not found")
			c.JSON(http.StatusNotFound, gin.H{"message": "rent not found"})
			return
		}

		mutation := struct {
			DeleteRent struct {
				ID graphql.Int `json:"id"`
			} `graphql:" delete_rentedBooks_by_pk(id: $id)"`
		}{}

		mutateVars := map[string]interface{}{
			"id": graphql.Int(rentId),
		}

		if err := client.Mutate(ctx, &mutation, mutateVars); err != nil {
			log.Println("unable to delete a rent", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to delete a rent", "details": err.Error()})
			return
		}

		if mutation.DeleteRent.ID == 0 {
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
			RentId int `json:"rentId"`
			BookId int `json:"bookId"`
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

		mutate := struct {
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
			} `graphql:"users(where: {wishlist: {bookId: {_eq: $bookId}}})"`
			Book struct {
				Title string `json:"title"`
			} `graphql:"books_by_pk(id: $bookId)"`
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
				} `graphql:" insert_notification_one(object: {id: 10, message: $message, userId: $userId})"`
			}{}
			notificationVars := map[string]interface{}{
				"userId":  graphql.Int(user.UserId),
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



func PaymentCallback() gin.HandlerFunc {
	return func(c *gin.Context) {
		var callbackData struct {
			Message string `json:"message"`
			Status  string `json:"status"`
			Data    struct {
				TxRef    string `json:"tx_ref"`
				Status   string `json:"status"`
				Amount   int    `json:"amount"`
				Currency string `json:"currency"`
			} `json:"data"`
		}

		// Bind the incoming callback data to the struct
		if err := c.ShouldBindJSON(&callbackData); err != nil {
			log.Println("Failed to parse callback data:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid callback data"})
			return
		}

		log.Println("Received payment callback:", callbackData)

		// Check if the payment status is "success"
		if callbackData.Data.Status != "success" {
			log.Println("Payment failed or not completed:", callbackData)
			c.JSON(http.StatusOK, gin.H{"message": "payment not successful"})
			return
		}

		// Proceed to update the payment status in the database
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		client := libs.SetupGraphqlClient()
		mutationVars := map[string]interface{}{
			"txRef": graphql.String(callbackData.Data.TxRef),
		}

		// Verify the payment before proceeding with status update
		isVerified, err := helpers.VerifyPayment(callbackData.Data.TxRef)
		if err != nil || !isVerified {
			log.Println("Payment verification failed:", err)
			c.JSON(http.StatusOK, gin.H{"message": "payment verification failed"})
			return
		}

		// Define the mutation for updating the payment status in the database
		type UpdatePaymentMutation struct {
			Status   graphql.String `graphql:"status"`
			TxRef    graphql.String `graphql:"tx_ref"`
		}

		var mutation struct {
			UpdatePayment struct {
				Returning []UpdatePaymentMutation `graphql:"returning"`
			} `graphql:"update_payments(where: {tx_ref: {_eq: $txRef}}, _set: {status: \"paid\"})"`
		}

		// Execute the mutation to update the payment status
		err = client.Mutate(ctx, &mutation, mutationVars)
		if err != nil {
			log.Println("Failed to update payment status:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update payment status"})
			return
		}

		// Log success and return a response
		log.Println("Payment status updated successfully for TxRef:", callbackData.Data.TxRef)
		c.JSON(http.StatusOK, gin.H{"message": "payment processed successfully"})
	}
}



func ProcessPayment() gin.HandlerFunc {
	return func(c *gin.Context) {
		client := libs.SetupGraphqlClient()
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var req requests.PaymentProcessRequest
		
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Println("Failed to bind JSON:", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid input", "details": err.Error()})
			return
		}
		if req.Input.TxRef == "" {
			log.Println("tx____ref is empty")
		}else{
			log.Println("tx____ref:" , req.Input.TxRef)
		}

		log.Println("tx____ref_id:" , req.Input.Id)
		// Verify the transaction reference
		isVerified, err := helpers.VerifyPayment(req.Input.TxRef)
		if err != nil || !isVerified {
			log.Println("Payment verification failed:", err)
			c.JSON(http.StatusOK, gin.H{"message": "payment verification failed" ,"details": err.Error()})
			return
		}

		// Update payment status in payments table
		mutationVars := map[string]interface{}{
			"id": graphql.Int(req.Input.Id),
		}

		type UpdatePaymentMutation struct {
			Status   graphql.String `graphql:"status"`
			TxRef    graphql.String `graphql:"tx_ref"`
		}

		var mutation struct {
			UpdatePayment struct {
				Returning []UpdatePaymentMutation `graphql:"returning"`
			} `graphql:"update_payments(where: {id: {_eq: $id}}, _set: {status: \"paid\"})"`
		}
		

		err = client.Mutate(ctx, &mutation, mutationVars)
		if err != nil {
			log.Println("Failed to update payment status:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update payment status"})
			return
		}

		log.Println("Payment status updated successfully for Payment ID:", req.Input.Id)
		log.Println("Payment status updated successfully for Payment txref:", req.Input.TxRef)
		res := models.ProcessPaymentOutput{
			Message:"payment processed successfully and status is updated",
			Status: "success",
			
		}
		c.JSON(http.StatusOK, res)
	}
}

