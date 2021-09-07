package random_test

import (
	"testing"

	"github.com/stevealexrs/Go-Libra/random"
)

func TestToken(t *testing.T) {
	token, err := random.Token20Byte()
	if err != nil {
		t.Error(err)
	}

	t.Log(token)
}
