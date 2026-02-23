package format

import (
	"encoding/binary"
	"errors"

	"github.com/amauribechtoldjr/msk/internal/domain"
	"github.com/amauribechtoldjr/msk/internal/wipe"
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

const (
	MSK_NAME_LENGTH_SIZE     = 2
	MSK_PASSWORD_LENGTH_SIZE = 2
)

var ErrCorruptedFile = errors.New("corrupted file")
var ErrUnsupportedFileVersion = errors.New("unsupported file version")

func getBufferLength(secret domain.Secret) int {
	return MSK_NAME_LENGTH_SIZE +
		len(secret.Name) +
		MSK_PASSWORD_LENGTH_SIZE +
		len(secret.Password)
}

func MarshalSecret(secret domain.Secret) []byte {
	bytesName := []byte(secret.Name)

	offset := 0
	buf := make([]byte, getBufferLength(secret))
	binary.BigEndian.PutUint16(buf[offset:], uint16(len(bytesName)))

	offset += MSK_NAME_LENGTH_SIZE

	copy(buf[offset:], []byte(secret.Name))

	offset += len(secret.Name)
	binary.BigEndian.PutUint16(buf[offset:], uint16(len(secret.Password)))

	offset += MSK_PASSWORD_LENGTH_SIZE

	copy(buf[offset:], []byte(secret.Password))

	return buf
}

func UnmarshalSecret(data []byte) (domain.Secret, error) {
	defer wipe.Bytes(data)

	secret := &domain.Secret{}
	offset := 0

	if len(data) < MSK_NAME_LENGTH_SIZE {
		return domain.Secret{}, ErrCorruptedFile
	}

	nameLen := int(binary.BigEndian.Uint16(data[offset:]))
	offset += MSK_NAME_LENGTH_SIZE

	if offset+nameLen > len(data) {
		return domain.Secret{}, ErrCorruptedFile
	}

	secret.Name = string(data[offset : offset+nameLen])
	offset += nameLen

	if offset+MSK_PASSWORD_LENGTH_SIZE > len(data) {
		return domain.Secret{}, ErrCorruptedFile
	}

	passLen := int(binary.BigEndian.Uint16(data[offset:]))
	offset += MSK_PASSWORD_LENGTH_SIZE

	if offset+passLen > len(data) {
		return domain.Secret{}, ErrCorruptedFile
	}

	secret.Password = make([]byte, passLen)
	copy(secret.Password, data[offset:offset+passLen])

	return *secret, nil
}

func MarshalFile(salt, nonce, data []byte) ([]byte, error) {
	if len(salt) != MSK_SALT_SIZE {
		return nil, errors.New("invalid salt size")
	}

	if len(nonce) != MSK_NONCE_SIZE {
		return nil, errors.New("invalid nonce size")
	}

	file := make([]byte, MSK_HEADER_SIZE+len(data))

	offset := 0
	copy(file[offset:], []byte(MSK_MAGIC_VALUE))

	offset += MSK_MAGIC_SIZE
	file[offset] = MSK_FILE_VERSION

	offset += MSK_VERSION_SIZE
	copy(file[offset:], salt)

	offset += MSK_SALT_SIZE
	copy(file[offset:], nonce)

	offset += MSK_NONCE_SIZE
	copy(file[offset:], data)

	return file, nil
}

func UnmarshalFile(data []byte) (salt, nonce, secret []byte, err error) {
	if len(data) < MSK_HEADER_SIZE {
		return nil, nil, nil, ErrCorruptedFile
	}

	if string(data[:MSK_MAGIC_SIZE]) != MSK_MAGIC_VALUE {
		return nil, nil, nil, ErrCorruptedFile
	}

	if len(data) <= MSK_MAGIC_SIZE {
		return nil, nil, nil, ErrCorruptedFile
	}

	if data[MSK_MAGIC_SIZE] != MSK_FILE_VERSION {
		return nil, nil, nil, ErrUnsupportedFileVersion
	}

	offset := MSK_MAGIC_SIZE + MSK_VERSION_SIZE

	salt = data[offset : offset+MSK_SALT_SIZE]
	offset += MSK_SALT_SIZE

	nonce = data[offset : offset+MSK_NONCE_SIZE]
	offset += MSK_NONCE_SIZE

	secret = data[offset:]

	return salt, nonce, secret, nil
}
