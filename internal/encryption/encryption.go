package encryption

import "github.com/amauribechtoldjr/msk/internal/domain"

const (
	MSK_MAGIC_VALUE   = "MSK"
	MSK_FILE_VERSION  = byte(1)

	MSK_MAGIC_SIZE    = 3
	MSK_VERSION_SIZE  = 1
	MSK_SALT_SIZE     = 16
	MSK_NONCE_SIZE    = 12
	MSK_HEADER_SIZE   = MSK_MAGIC_SIZE + MSK_VERSION_SIZE + MSK_SALT_SIZE + MSK_NONCE_SIZE
)

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