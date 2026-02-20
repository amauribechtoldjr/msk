package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"errors"

	"github.com/amauribechtoldjr/msk/internal/domain"
	"github.com/amauribechtoldjr/msk/internal/wipe"
	"github.com/awnumar/memguard"
)

const (
	MSK_MAGIC_VALUE  = "MSK"
	MSK_FILE_VERSION = byte(1)

	MSK_MAGIC_SIZE   = 3
	MSK_VERSION_SIZE = 1
	MSK_SALT_SIZE    = 16
	MSK_NONCE_SIZE   = 12
	MSK_HEADER_SIZE  = MSK_MAGIC_SIZE + MSK_VERSION_SIZE + MSK_SALT_SIZE + MSK_NONCE_SIZE
)

var ErrDecryption = errors.New("decryption failed")
var ErrCorruptedFile = errors.New("corrupted file")
var ErrUnsupportedFileVersion = errors.New("unsupported file version")

type Encryption interface {
	Encrypt(secret domain.Secret) (domain.EncryptedSecret, error)
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
	if len(cipherData) < MSK_HEADER_SIZE {
		return domain.Secret{}, ErrCorruptedFile
	}

	if string(cipherData[:MSK_MAGIC_SIZE]) != MSK_MAGIC_VALUE {
		return domain.Secret{}, ErrCorruptedFile
	}

	if cipherData[MSK_MAGIC_SIZE] != MSK_FILE_VERSION {
		return domain.Secret{}, ErrUnsupportedFileVersion
	}

	offset := MSK_MAGIC_SIZE + MSK_VERSION_SIZE

	salt := cipherData[offset : offset+MSK_SALT_SIZE]
	offset += MSK_SALT_SIZE

	nonce := cipherData[offset : offset+MSK_NONCE_SIZE]
	offset += MSK_NONCE_SIZE

	cipherText := cipherData[offset:]

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

	plaintext, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return domain.Secret{}, ErrDecryption
	}
	defer wipe.Bytes(plaintext)

	var s domain.Secret
	if err := json.Unmarshal(plaintext, &s); err != nil {
		return domain.Secret{}, err
	}

	return s, nil
}

func (a *ArgonCrypt) Encrypt(secret domain.Secret) (domain.EncryptedSecret, error) {
	salt, err := randomBytes(MSK_SALT_SIZE)
	if err != nil {
		return domain.EncryptedSecret{}, err
	}

	if a.mk == nil {
		return domain.EncryptedSecret{}, errors.New("failed to load master key")
	}

	lockedBuffer, err := a.mk.Open()
	if err != nil {
		return domain.EncryptedSecret{}, err
	}
	defer lockedBuffer.Destroy()

	key, err := getArgonDeriveKey(lockedBuffer.Bytes(), salt)
	if err != nil {
		return domain.EncryptedSecret{}, err
	}
	defer wipe.Bytes(key)

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
	defer wipe.Bytes(plaintext)

	return domain.EncryptedSecret{
		Data:  cipherText,
		Salt:  [MSK_SALT_SIZE]byte(salt),
		Nonce: [MSK_NONCE_SIZE]byte(nonce),
	}, nil
}
