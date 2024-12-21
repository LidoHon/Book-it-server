package helpers

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

func HashedPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Panic(err)
	}
	return string(hashedPassword)
}

func VerifyPassword(userPassword string, hashedpassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedpassword), []byte(userPassword))
	check := true
	msg := ""
	if err != nil {
		msg = "login or password incorrect"
		check = false

	}
	return check, msg
}
