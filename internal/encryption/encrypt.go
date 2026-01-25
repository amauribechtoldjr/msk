package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"

	"github.com/amauribechtoldjr/msk/internal/domain"
)

func (a *ArgonCrypt) Encrypt(secret domain.Secret) (domain.EncryptedSecret, error) {
	salt, err := randomBytes(MSK_SALT_SIZE)
	if err != nil {
		return domain.EncryptedSecret{}, err
	}

	key := getArgonDeriveKey(a.mk, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return domain.EncryptedSecret{}, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return domain.EncryptedSecret{}, err
	}

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
		Data:  cipherText,
		Salt:  [MSK_SALT_SIZE]byte(salt),
		Nonce: [MSK_NONCE_SIZE]byte(nonce),
	}, nil
}
