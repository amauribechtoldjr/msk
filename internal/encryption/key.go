package encryption

import (
	"errors"

	"golang.org/x/crypto/argon2"
)

var ErrInvalidSalt = errors.New("invalid salt size")
var ErrInvalidPass = errors.New("invalid master pass")

func getArgonDeriveKey(password, salt []byte) ([]byte, error) {
	if len(salt) != MSK_SALT_SIZE {
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
