package random

import (
	"crypto/rand"
	"encoding/hex"
)


const otpChars = "1234567890"

func OTP() (string, error) {
	b := make([]byte, 6)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	for i := range b {
		b[i] = otpChars[int(b[i]) % len(otpChars)]
	}

	return string(b), nil
}

func generateSecureToken(length int) (string, error) {
	b := make([]byte, length)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// 16 byte is good enough for most thing, very small chance of collision
func Token16Byte() (string, error) {
	return generateSecureToken(16)
}
// If collision is an issue, 20 byte has virtually no chance of collision
func Token20Byte() (string, error) {
	return generateSecureToken(20)
}
