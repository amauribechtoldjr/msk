package vault

import (
	"crypto/sha256"
	"errors"
)

// DeriveSessionToken hashes the given token into a 32-byte key suitable for
// AES-GCM encryption of the session file.
func DeriveSessionToken(token []byte) ([]byte, error) {
	if len(token) == 0 {
		return nil, errors.New("invalid session token")
	}

	sum := sha256.Sum256(token)
	key := make([]byte, 32)

	copy(key, sum[:])

	return key, nil
}
