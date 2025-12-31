package file_crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"io"

	u "github.com/amauribechtoldjr/msk/utils"
	"golang.org/x/crypto/pbkdf2"
)

var p = u.Panic

func Encrypt(data []byte, masterKey []byte) []byte {
	nonce := make([]byte, 12)

	_, err := io.ReadFull(rand.Reader, nonce);
	p(err)

	derivedKey := pbkdf2.Key(masterKey, nonce, 4096, 32, sha1.New)
	block, err := aes.NewCipher(derivedKey)
	p(err);

	aesGcm, err := cipher.NewGCM(block)
	p(err)

	cipherText := aesGcm.Seal(nil, nonce, data, nil)
	cipherText = append(cipherText, nonce...)

	return cipherText
}

func Decrypt(cipherData []byte, masterKey []byte) []byte {
	salt := cipherData[len(cipherData)-12:]
	encodedSalt := hex.EncodeToString(salt)
	nonce, err := hex.DecodeString(encodedSalt)
	p(err)

	derivedKey := pbkdf2.Key(masterKey, nonce, 4096, 32, sha1.New)
	block, err := aes.NewCipher(derivedKey)
	p(err)

	aesCgm, err := cipher.NewGCM(block)
	p(err)

	plainText, err := aesCgm.Open(nil, nonce, cipherData[:len(cipherData)-12], nil)
	p(err)

	return plainText
}
