package vault

import (
	"errors"

	"github.com/amauribechtoldjr/msk/internal/meta"
	"golang.org/x/crypto/argon2"
)

var ErrInvalidSalt = errors.New("invalid salt size")
var ErrInvalidPass = errors.New("invalid master pass")

// SecretKeyDeriver derives an AES-256 encryption key from the user's
// master password using Argon2id. This is intentionally slow and
// memory-hard to resist brute-force attacks on low-entropy passwords.
type SecretKeyDeriver struct{}

func (s *SecretKeyDeriver) getSecretKey(password, salt []byte) ([]byte, error) {
	if len(salt) != meta.MSK_SALT_SIZE {
		return nil, ErrInvalidSalt
	}

	if len(password) == 0 {
		return nil, ErrInvalidPass
	}

	return argon2.IDKey(
		password,
		salt,
		6,
		128*1024,
		4,
		32,
	), nil
}
