package vault

import (
	"errors"

	"github.com/amauribechtoldjr/msk/internal/domain"
	"github.com/amauribechtoldjr/msk/internal/format"
	"github.com/amauribechtoldjr/msk/internal/gcm"
	"github.com/amauribechtoldjr/msk/internal/meta"
	"github.com/amauribechtoldjr/msk/internal/session"

	"github.com/amauribechtoldjr/msk/internal/wipe"
	"github.com/awnumar/memguard"
)

var ErrDecryption = errors.New("decryption failed")

type Vault interface {
	Encrypt([]byte) ([]byte, error)
	DecryptSecret([]byte) (domain.Secret, error)
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

func (a *vault) DecryptSecret(cipherData []byte) (domain.Secret, error) {
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

	fileBytes, err := gcm.OpenGCM(nonce, key, data)
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

func (a *vault) Encrypt(fileBytes []byte) ([]byte, error) {
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

	deriver := &SecretKeyDeriver{}
	key, err := deriver.getSecretKey(lockedBuffer.Bytes(), salt)
	if err != nil {
		return nil, err
	}
	defer wipe.Bytes(key)

	sealedGCM, err := gcm.SealGCM(key, fileBytes)
	if err != nil {
		return nil, err
	}
	defer wipe.Bytes(fileBytes)

	return format.MarshalFile(salt, sealedGCM.Nonce, sealedGCM.CipherData)
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
