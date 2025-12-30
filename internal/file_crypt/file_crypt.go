package file_crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"io"

	u "github.com/amauribechtoldjr/msk/utils"
	"golang.org/x/crypto/pbkdf2"
)

var c = u.Check

func Encrypt(data []byte, masterKey []byte) []byte {
	nonce := make([]byte, 12)

	_, err := io.ReadFull(rand.Reader, nonce);
	c(err)

	derivedKey := pbkdf2.Key(masterKey, nonce, 4096, 32, sha1.New)
	block, err := aes.NewCipher(derivedKey)
	c(err);

	aesGcm, err :=cipher.NewGCM(block)
	c(err)

	cipherText := aesGcm.Seal(nil, nonce, data, nil)
	cipherText = append(cipherText, nonce...)

	return cipherText
}

func Decrypt() {

}
