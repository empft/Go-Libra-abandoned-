package random

import "crypto/rand"

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