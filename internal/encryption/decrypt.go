package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"errors"

	"github.com/amauribechtoldjr/msk/internal/domain"
)

var ErrDecrypt = errors.New("failed to decrypt file")

func cleanupByte(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func (a *ArgonCrypt) Decrypt(cipherData []byte) (domain.Secret, error) {
	if len(cipherData) < MSK_HEADER_SIZE {
		return domain.Secret{}, errors.New("invalid or corrupted file")
	}

	if string(cipherData[:MSK_MAGIC_SIZE]) != MSK_MAGIC_VALUE {
		return domain.Secret{}, errors.New("invalid file format")
	}

	if cipherData[MSK_MAGIC_SIZE] != MSK_FILE_VERSION {
		return domain.Secret{}, errors.New("unsupported file version")
	}

	offset := MSK_MAGIC_SIZE + MSK_VERSION_SIZE

	salt := cipherData[offset : offset+MSK_SALT_SIZE]
	offset += MSK_SALT_SIZE

	nonce := cipherData[offset : offset+MSK_NONCE_SIZE]
	offset += MSK_NONCE_SIZE

	cipherText := cipherData[offset:]

	// ---- Key derivation ----
	key := getArgonDeriveKey(a.mk, salt)
	defer cleanupByte(key)

	// ---- Decryption ----
	block, err := aes.NewCipher(key)
	if err != nil {
		return domain.Secret{}, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return domain.Secret{}, err
	}

	plaintext, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return domain.Secret{}, errors.New("authentication failed")
	}
	defer cleanupByte(plaintext)

	// ---- Decode ----
	var s domain.Secret
	if err := json.Unmarshal(plaintext, &s); err != nil {
		return domain.Secret{}, err
	}

	return s, nil
}
