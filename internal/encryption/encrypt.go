package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"errors"

	"github.com/amauribechtoldjr/msk/internal/domain"
)

var ErrEncrypt = errors.New("failed to encrypt file")



func (a *ArgonCrypt) Encrypt(secret domain.Secret) (domain.EncryptedSecret, error) {
	salt, err := randomBytes(MSK_SALT_SIZE)
	if err != nil {
		return domain.EncryptedSecret{}, err
	}

	key := getArgonDeriveKey(a.mk, salt)

	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)

	nonce, err := randomBytes(MSK_NONCE_SIZE)
	if err != nil {
		return domain.EncryptedSecret{}, err
	}

	plaintext, err := json.Marshal(secret)
	if err != nil {
		return domain.EncryptedSecret{}, err
	}
	
	cipherText := gcm.Seal(nil, nonce, plaintext, nil)

	return domain.EncryptedSecret{
		Data: cipherText, 
		Salt: [MSK_SALT_SIZE]byte(salt), 
		Nonce: [MSK_NONCE_SIZE]byte(nonce),
	}, nil
}