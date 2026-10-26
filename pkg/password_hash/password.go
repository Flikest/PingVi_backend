package passwordhash

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func Hashing(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func Checking(password string, hashPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password))
	fmt.Println(err)
	return err == nil
}
