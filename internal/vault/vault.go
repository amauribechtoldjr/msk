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
	LoadSession(token []byte, nonce []byte, cipherData []byte) ([]byte, error)
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

// DestroyMK wipes all memguard-managed memory and releases the master key reference.
func (ac *MSKVault) DestroyMK() {
	memguard.Purge()
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

	deriver := &SecretKeyDeriver{}
	key, err := deriver.getSecretKey(lockedBuffer.Bytes(), salt)
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
	salt, err := format.RandomBytes(format.MSK_SALT_SIZE)
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

	deriver := &SecretKeyDeriver{}
	key, err := deriver.getSecretKey(lockedBuffer.Bytes(), salt)
	if err != nil {
		return nil, err
	}
	defer wipe.Bytes(key)

	fileBytes := format.MarshalSecret(secret)

	nonce, cipherData, err := sealGCM(key, fileBytes)
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

	hasher := &TokenHasher{}
	key, err := hasher.getSessionToken(token)
	if err != nil {
		return nil, nil, err
	}
	defer wipe.Bytes(key)

	nonce, cipherData, err = sealGCM(key, lockedBuffer.Bytes())
	if err != nil {
		return nil, nil, err
	}

	return nonce, cipherData, nil
}

func (a *MSKVault) LoadSession(token []byte, nonce []byte, cipherData []byte) ([]byte, error) {
	hasher := &TokenHasher{}
	key, err := hasher.getSessionToken(token)
	if err != nil {
		return nil, err
	}
	defer wipe.Bytes(key)

	mk, err := openGCM(nonce, key, cipherData)
	if err != nil {
		return nil, ErrDecryption
	}

	return mk, nil
}

func sealGCM(key []byte, fileBytes []byte) (nonce []byte, cipherData []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce, err = format.RandomBytes(gcm.NonceSize())
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
