package entity

import scrypt "github.com/elithrar/simple-scrypt"

func generateHash(password string) ([]byte, error) {
	return scrypt.GenerateFromPassword([]byte(password), scrypt.DefaultParams)
}