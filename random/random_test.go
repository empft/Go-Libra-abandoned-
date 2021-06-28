package random

import "testing"

func TestToken(t *testing.T) {
	token, err := Token20Byte()
	if err != nil {
		t.Error(err)
	}

	t.Log(token)
}
