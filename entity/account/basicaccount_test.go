package entity

import "testing"

func TestHash(t *testing.T) {
	hash, err := generateHash("Password")
	if err != nil {
		t.Error(err)
	}
	t.Log(hash)
}