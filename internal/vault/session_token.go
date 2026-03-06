package vault

import (
	"crypto/sha256"
	"errors"
)

// TokenHasher derives an AES-256 key from a high-entropy session token
// using SHA-256. The input must be cryptographically random (e.g. 32 bytes
// from crypto/rand).
type TokenHasher struct{}

// getSessionToken hashes the given token into a 32-byte key suitable for
// AES-GCM encryption of the session file.
func (t *TokenHasher) getSessionToken(token []byte) ([]byte, error) {
	if len(token) == 0 {
		return nil, errors.New("invalid session token")
	}

	sum := sha256.Sum256(token)
	key := make([]byte, 32)

	copy(key, sum[:])

	return key, nil
}
