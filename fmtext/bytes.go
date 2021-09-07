package fmtext

import (
	"fmt"
)

// returns the formatted byte string
func Byte(size int, decimalPlaces int) string {
	const multiplier = 1000
	units := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}

	current := float64(size)
	for i, v := range units {
		if current < multiplier || i == (len(units) - 1) {
			return fmt.Sprintf("%.*f %s", decimalPlaces, current, v)
		}
		current /= multiplier
	}
	return "Inaccessible Region"
}
