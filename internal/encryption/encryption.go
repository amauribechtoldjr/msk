package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"errors"

	"github.com/amauribechtoldjr/msk/internal/domain"
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
}

type ArgonCrypt struct {
	mk []byte
}

func NewArgonCrypt() *ArgonCrypt {
	return &ArgonCrypt{}
}

func (ac *ArgonCrypt) ConfigMK(mk []byte) {
	ac.mk = mk
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

	key, err := getArgonDeriveKey(a.mk, salt)
	if err != nil {
		return domain.Secret{}, err
	}

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

	key, err := getArgonDeriveKey(a.mk, salt)
	if err != nil {
		return domain.EncryptedSecret{}, err
	}

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
