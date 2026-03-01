package vault

import (
	"crypto/sha256"
	"errors"
)

type SHA256 struct{}

func (a *SHA256) DeriveKey(password []byte) ([]byte, error) {
	if len(password) == 0 {
		return nil, errors.New("invalid master pass")
	}

	sum := sha256.Sum256(password)
	key := make([]byte, 32)

	copy(key, sum[:])

	return key, nil
}
