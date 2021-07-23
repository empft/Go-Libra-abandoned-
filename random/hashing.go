package random

import scrypt "github.com/elithrar/simple-scrypt"

func GenerateHash(password string) ([]byte, error) {
	return scrypt.GenerateFromPassword([]byte(password), scrypt.DefaultParams)
}

func CompareHashAndPassword(hash []byte, password string) error {
	return scrypt.CompareHashAndPassword(hash, []byte(password))
}