package vault

import (
	"errors"

	"github.com/amauribechtoldjr/msk/internal/format"
	"github.com/amauribechtoldjr/msk/internal/gcm"
	"github.com/amauribechtoldjr/msk/internal/meta"
	"github.com/amauribechtoldjr/msk/internal/session"

	"github.com/amauribechtoldjr/msk/internal/wipe"
	"github.com/awnumar/memguard"
)

var ErrDecryption = errors.New("decryption failed")

type Vault interface {
	Encrypt([]byte) (*gcm.SaltedGCM, error)
	Decrypt(salt, nonce, data []byte) ([]byte, error)
	ConfigMK([]byte)
	DestroyMK()
	CreateSession(token []byte) (*gcm.SealedCGM, error)
	LoadSession(bs *session.BinarySession) error
}

type vault struct {
	mk *memguard.Enclave
}

func NewMSKVault() Vault {
	return &vault{}
}

func (ac *vault) ConfigMK(mk []byte) {
	buffer := memguard.NewBufferFromBytes(mk)
	ac.mk = buffer.Seal()
}

// DestroyMK wipes all memguard-managed memory and releases the master key reference.
func (ac *vault) DestroyMK() {
	memguard.Purge()
	ac.mk = nil
}

func (a *vault) Decrypt(salt, nonce, data []byte) ([]byte, error) {
	if a.mk == nil {
		return nil, errors.New("failed to load master key")
	}

	lockedBuffer, err := a.mk.Open()
	if err != nil {
		return nil, err
	}
	defer lockedBuffer.Destroy()

	key, err := DeriveArgonKey(lockedBuffer.Bytes(), salt)
	if err != nil {
		return nil, err
	}
	defer wipe.Bytes(key)

	fileBytes, err := gcm.OpenGCM(nonce, key, data)
	if err != nil {
		return nil, ErrDecryption
	}

	return fileBytes, nil
}

func (a *vault) Encrypt(fileBytes []byte) (*gcm.SaltedGCM, error) {
	salt, err := format.RandomBytes(meta.MSK_SALT_SIZE)
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

	key, err := DeriveArgonKey(lockedBuffer.Bytes(), salt)
	if err != nil {
		return nil, err
	}
	defer wipe.Bytes(key)

	sealedGCM, err := gcm.SealGCM(key, fileBytes)
	if err != nil {
		return nil, err
	}
	defer wipe.Bytes(fileBytes)

	return &gcm.SaltedGCM{Nonce: sealedGCM.Nonce, Salt: salt, CipherData: sealedGCM.CipherData}, nil
}

func (a *vault) CreateSession(token []byte) (*gcm.SealedCGM, error) {
	if a.mk == nil {
		return nil, errors.New("failed to load master key")
	}

	lockedBuffer, err := a.mk.Open()
	if err != nil {
		return nil, err
	}
	defer lockedBuffer.Destroy()

	key, err := DeriveSessionToken(token)
	if err != nil {
		return nil, err
	}
	defer wipe.Bytes(key)

	sealedGCM, err := gcm.SealGCM(key, lockedBuffer.Bytes())
	if err != nil {
		return nil, err
	}

	return sealedGCM, nil
}

func (a *vault) LoadSession(bs *session.BinarySession) error {
	key, err := DeriveSessionToken(bs.Token)
	if err != nil {
		return err
	}
	defer wipe.Bytes(key)

	mk, err := gcm.OpenGCM(bs.Nonce[:], key, bs.CipherData)
	if err != nil {
		return ErrDecryption
	}

	a.ConfigMK(mk)

	return nil
}
