package random

import "testing"

func TestHash(t *testing.T) {
	hash, err := GenerateHash("Password")
	if err != nil {
		t.Error(err)
	}
	t.Log(hash)
}