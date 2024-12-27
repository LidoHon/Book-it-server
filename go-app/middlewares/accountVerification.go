package middlewares

// import (
// "net/http"
// "github.com/golang-jwt/jwt/v4"
// "strings"
// )

// func ValidateAccessToken(next http.Handler) http.Handler {
// return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// authHeader := r.Header.Get("Authorization")
// if authHeader == "" {
// http.Error(w, "Missing token", http.StatusUnauthorized)
// return
// }

// tokenString := strings.Split(authHeader, " ")[1]

// token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// return jwtSecret, nil
// })

// if err != nil || !token.Valid {
// http.Error(w, "Invalid token", http.StatusUnauthorized)
// return
// }

// next.ServeHTTP(w, r)
// })
// }
