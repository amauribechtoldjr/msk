package encryption

import "golang.org/x/crypto/argon2"

func getArgonDeriveKey(password, salt []byte) []byte {
	return argon2.IDKey(
		password,
		salt,
		3,
		64*1024,
		4,
		32,
	)
}