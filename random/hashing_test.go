package random_test

import (
	"testing"

	"github.com/stevealexrs/Go-Libra/random"
)

func TestHash(t *testing.T) {
	hash, err := random.GenerateHash("Password")
	if err != nil {
		t.Error(err)
	}
	t.Log(hash)
}