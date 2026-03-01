package vault

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

type Vault interface {
	EncryptSecret(secret domain.Secret) ([]byte, error)
	DecryptSecret(cipherData []byte) (domain.Secret, error)
	ConfigMK(mk []byte)
	DestroyMK()
	CreateSession(token []byte) (nonce []byte, cipherData []byte, err error)
}

type MSKVault struct {
	mk *memguard.Enclave
}

func NewMSKVault() *MSKVault {
	return &MSKVault{}
}

func (ac *MSKVault) ConfigMK(mk []byte) {
	buffer := memguard.NewBufferFromBytes(mk)
	ac.mk = buffer.Seal()
}

func (ac *MSKVault) DestroyMK() {
	ac.mk = nil
}

func (a *MSKVault) DecryptSecret(cipherData []byte) (domain.Secret, error) {
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

	argon2 := &Argon2{}
	key, err := argon2.DeriveKey(lockedBuffer.Bytes(), salt)
	if err != nil {
		return domain.Secret{}, err
	}
	defer wipe.Bytes(key)

	fileBytes, err := openGCM(nonce, key, data)
	if err != nil {
		return domain.Secret{}, ErrDecryption
	}
	defer wipe.Bytes(fileBytes)

	secret, err := format.UnmarshalSecret(fileBytes)
	if err != nil {
		return domain.Secret{}, err
	}

	return secret, nil
}

func (a *MSKVault) EncryptSecret(secret domain.Secret) ([]byte, error) {
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

	argon2 := &Argon2{}
	key, err := argon2.DeriveKey(lockedBuffer.Bytes(), salt)
	if err != nil {
		return nil, err
	}
	defer wipe.Bytes(key)

	fileBytes := format.MarshalSecret(secret)

	nonce, cipherData, err := sealGCM(format.MSK_NONCE_SIZE, key, fileBytes)
	if err != nil {
		return nil, err
	}
	defer wipe.Bytes(fileBytes)
	defer wipe.Bytes(cipherData)

	return format.MarshalFile(salt, nonce, cipherData)
}

func (a *MSKVault) CreateSession(token []byte) (nonce []byte, cipherData []byte, err error) {
	if a.mk == nil {
		return nil, nil, errors.New("failed to load master key")
	}

	lockedBuffer, err := a.mk.Open()
	if err != nil {
		return nil, nil, err
	}
	defer lockedBuffer.Destroy()

	sha256 := &SHA256{}
	key, err := sha256.DeriveKey(token)
	if err != nil {
		return nil, nil, err
	}
	defer wipe.Bytes(key)

	nonce, cipherData, err = sealGCM(SESSION_NONCE_SIZE, key, lockedBuffer.Bytes())
	if err != nil {
		return nil, nil, err
	}

	return nonce, cipherData, nil
}

func sealGCM(nonceSize int, key []byte, fileBytes []byte) (nonce []byte, cipherData []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce, err = randomBytes(nonceSize)
	if err != nil {
		return nil, nil, err
	}

	return nonce, gcm.Seal(nil, nonce, fileBytes, nil), nil
}

func openGCM(nonce, key, data []byte) ([]byte, error) {
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
