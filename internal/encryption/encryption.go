package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"

	"github.com/amauribechtoldjr/msk/internal/domain"
	"github.com/amauribechtoldjr/msk/internal/format"
	"github.com/amauribechtoldjr/msk/internal/wipe"
	"github.com/awnumar/memguard"
)

var ErrDecryption = errors.New("decryption failed")

type Encryption interface {
	Encrypt(secret domain.Secret) ([]byte, error)
	Decrypt(cipherData []byte) (domain.Secret, error)
	ConfigMK(mk []byte)
	DestroyMK()
}

type ArgonCrypt struct {
	mk *memguard.Enclave
}

func NewArgonCrypt() *ArgonCrypt {
	return &ArgonCrypt{}
}

func (ac *ArgonCrypt) ConfigMK(mk []byte) {
	buffer := memguard.NewBufferFromBytes(mk)
	ac.mk = buffer.Seal()
}

func (ac *ArgonCrypt) DestroyMK() {
	ac.mk = nil
}

func (a *ArgonCrypt) Decrypt(cipherData []byte) (domain.Secret, error) {
	salt, nonce, data, err := format.UnmarshalFile(cipherData)
	if err != nil {
		return domain.Secret{}, err
	}

	if a.mk == nil {
		return domain.Secret{}, errors.New("failed to load master key")
	}

	lockedBuffer, err := a.mk.Open()
	if err != nil {
		return domain.Secret{}, err
	}
	defer lockedBuffer.Destroy()

	key, err := getArgonDeriveKey(lockedBuffer.Bytes(), salt)
	if err != nil {
		return domain.Secret{}, err
	}
	defer wipe.Bytes(key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return domain.Secret{}, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return domain.Secret{}, err
	}

	fileBytes, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return domain.Secret{}, ErrDecryption
	}
	defer wipe.Bytes(fileBytes)

	secret := format.UnmarshalSecret(fileBytes)

	return secret, nil
}

func (a *ArgonCrypt) Encrypt(secret domain.Secret) ([]byte, error) {
	salt, err := randomBytes(format.MSK_SALT_SIZE)
	if err != nil {
		return nil, err
	}

	if a.mk == nil {
		return nil, errors.New("failed to load master key")
	}

	lockedBuffer, err := a.mk.Open()
	if err != nil {
		return nil, err
	}
	defer lockedBuffer.Destroy()

	key, err := getArgonDeriveKey(lockedBuffer.Bytes(), salt)
	if err != nil {
		return nil, err
	}
	defer wipe.Bytes(key)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce, err := randomBytes(format.MSK_NONCE_SIZE)
	if err != nil {
		return nil, err
	}

	fileBytes := format.MarshalSecret(secret)

	cipherBytes := gcm.Seal(nil, nonce, fileBytes, nil)
	defer wipe.Bytes(fileBytes)

	return format.MarshalFile(salt, nonce, cipherBytes)
}
