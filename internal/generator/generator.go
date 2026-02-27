package generator

import (
	"crypto/rand"
	"math/big"
)

const (
	alphanumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	symbols      = "!@#$%^&*()-_=+[]{}|;:,.<>?"
)

func GeneratePassword(length int, noSymbols bool) ([]byte, error) {
	if length <= 0 {
		length = 16
	}

	charset := alphanumeric
	if !noSymbols {
		charset += symbols
	}

	password := make([]byte, length)
	for i := range password {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return nil, err
		}
		password[i] = charset[idx.Int64()]
	}

	return password, nil
}
