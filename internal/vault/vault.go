package vault

import (
	"errors"

	"github.com/amauribechtoldjr/msk/internal/format"
	"github.com/amauribechtoldjr/msk/internal/gcm"
	"github.com/amauribechtoldjr/msk/internal/meta"
	"github.com/amauribechtoldjr/msk/internal/prompt"
	"github.com/amauribechtoldjr/msk/internal/session"

	"github.com/amauribechtoldjr/msk/internal/wipe"
	"github.com/awnumar/memguard"
)

var ErrDecryption = errors.New("decryption failed")

type Vault interface {
	Encrypt([]byte) (*gcm.SaltedGCM, error)
	Decrypt(salt, nonce, data []byte) ([]byte, error)
	DestroyMK()
	CreateSession(token []byte) (*gcm.SealedCGM, error)
	LoadSession(bs *session.BinarySession) error
	LoadMK() error
}

type vault struct {
	mk *memguard.Enclave
}

func NewVault() Vault {
	return &vault{}
}

func NewVaultWithMK(mk []byte) Vault {
	v := &vault{}
	v.configMK(mk)
	return v
}

func (v *vault) configMK(mk []byte) {
	buffer := memguard.NewBufferFromBytes(mk)
	v.mk = buffer.Seal()
}

func (v *vault) DestroyMK() {
	memguard.Purge()
	v.mk = nil
}

func (v *vault) withMk(fn func(mk []byte) error) error {
	if v.mk == nil {
		return errors.New("failed to load master key")
	}

	lockedBuffer, err := v.mk.Open()
	if err != nil {
		return err
	}
	defer lockedBuffer.Destroy()

	return fn(lockedBuffer.Bytes())
}

func (v *vault) Decrypt(salt, nonce, data []byte) ([]byte, error) {
	var fileBytes []byte

	err := v.withMk(func(mk []byte) error {
		key, err := DeriveArgonKey(mk, salt)
		if err != nil {
			return err
		}
		defer wipe.Bytes(key)

		fileBytes, err = gcm.OpenGCM(nonce, key, data)
		if err != nil {
			return ErrDecryption
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}

func (v *vault) Encrypt(fileBytes []byte) (*gcm.SaltedGCM, error) {
	salt, err := format.RandomBytes(meta.MSK_SALT_SIZE)
	if err != nil {
		return nil, err
	}

	var sealedGCM *gcm.SealedCGM

	err = v.withMk(func(mk []byte) error {
		key, err := DeriveArgonKey(mk, salt)
		if err != nil {
			return err
		}
		defer wipe.Bytes(key)

		sealedGCM, err = gcm.SealGCM(key, fileBytes)
		if err != nil {
			return err
		}
		defer wipe.Bytes(fileBytes)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &gcm.SaltedGCM{Nonce: sealedGCM.Nonce, Salt: salt, CipherData: sealedGCM.CipherData}, nil
}

func (v *vault) CreateSession(token []byte) (*gcm.SealedCGM, error) {
	var sealedGCM *gcm.SealedCGM

	err := v.withMk(func(mk []byte) error {
		key, err := DeriveSessionToken(token)
		if err != nil {
			return err
		}
		defer wipe.Bytes(key)

		sealedGCM, err = gcm.SealGCM(key, mk)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return sealedGCM, nil
}

func (v *vault) LoadSession(bs *session.BinarySession) error {
	key, err := DeriveSessionToken(bs.Token)
	if err != nil {
		return err
	}
	defer wipe.Bytes(key)

	mk, err := gcm.OpenGCM(bs.Nonce[:], key, bs.CipherData)
	if err != nil {
		return ErrDecryption
	}

	v.configMK(mk)

	return nil
}

func (v *vault) LoadMK() error {
	mk, err := prompt.ReadMasterPassword(false)
	if err != nil {
		return err
	}
	defer wipe.Bytes(mk)
	v.configMK(mk)
	return nil
}
