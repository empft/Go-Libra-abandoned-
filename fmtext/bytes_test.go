package fmtext_test

import (
	"testing"

	"github.com/stevealexrs/Go-Libra/fmtext"
)

func TestByte(t *testing.T) {
	if res := fmtext.Byte(1000000, 0); res != "1 MB" {
		t.Errorf("expect 1 MB, got %v", res)
	}

	if res := fmtext.Byte(2300, 1); res != "2.3 kB" {
		t.Errorf("expect 2.3 kB, got %v", res)
	}

	if res := fmtext.Byte(2330000000000000000, 2); res != "2.33 EB" {
		t.Errorf("expect 2.33 EB, got %v", res)
	}
}
