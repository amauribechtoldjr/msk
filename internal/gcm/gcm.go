package gcm

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/amauribechtoldjr/msk/internal/format"
)

type SealedCGM struct {
	Nonce      []byte
	CipherData []byte
}

func SealGCM(key []byte, fileBytes []byte) (*SealedCGM, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce, err := format.RandomBytes(gcm.NonceSize())
	if err != nil {
		return nil, err
	}

	return &SealedCGM{nonce, gcm.Seal(nil, nonce, fileBytes, nil)}, nil
}

func OpenGCM(nonce, key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	fileBytes, err := gcm.Open(nil, nonce, data, nil)

	return fileBytes, err
}
