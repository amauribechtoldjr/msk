package format

import (
	"encoding/binary"
	"errors"

	"github.com/amauribechtoldjr/msk/internal/domain"
	"github.com/amauribechtoldjr/msk/internal/meta"
)

var ErrCorruptedFile = errors.New("corrupted file")
var ErrUnsupportedFileVersion = errors.New("unsupported file version")

func getBufferLength(secret domain.Secret) int {
	return meta.SECRET_NAME_LENGTH_SIZE +
		len(secret.Name) +
		meta.SECRET_PASSWORD_LENGTH_SIZE +
		len(secret.Password)
}

func MarshalSecret(secret domain.Secret) []byte {
	bytesName := []byte(secret.Name)

	offset := 0
	buf := make([]byte, getBufferLength(secret))
	binary.BigEndian.PutUint16(buf[offset:], uint16(len(bytesName)))

	offset += meta.SECRET_NAME_LENGTH_SIZE

	copy(buf[offset:], []byte(secret.Name))

	offset += len(secret.Name)
	binary.BigEndian.PutUint16(buf[offset:], uint16(len(secret.Password)))

	offset += meta.SECRET_PASSWORD_LENGTH_SIZE

	copy(buf[offset:], []byte(secret.Password))

	return buf
}

func UnmarshalSecret(data []byte) (domain.Secret, error) {
	secret := &domain.Secret{}
	offset := 0

	if len(data) < meta.SECRET_NAME_LENGTH_SIZE {
		return domain.Secret{}, ErrCorruptedFile
	}

	nameLen := int(binary.BigEndian.Uint16(data[offset:]))
	offset += meta.SECRET_NAME_LENGTH_SIZE

	if offset+nameLen > len(data) {
		return domain.Secret{}, ErrCorruptedFile
	}

	secret.Name = string(data[offset : offset+nameLen])
	offset += nameLen

	if offset+meta.SECRET_PASSWORD_LENGTH_SIZE > len(data) {
		return domain.Secret{}, ErrCorruptedFile
	}

	passLen := int(binary.BigEndian.Uint16(data[offset:]))
	offset += meta.SECRET_PASSWORD_LENGTH_SIZE

	if offset+passLen > len(data) {
		return domain.Secret{}, ErrCorruptedFile
	}

	secret.Password = make([]byte, passLen)
	copy(secret.Password, data[offset:offset+passLen])

	return *secret, nil
}

func MarshalFile(salt, nonce, data []byte) ([]byte, error) {
	if len(salt) != meta.MSK_SALT_SIZE {
		return nil, errors.New("invalid salt size")
	}

	if len(nonce) != meta.MSK_NONCE_SIZE {
		return nil, errors.New("invalid nonce size")
	}

	file := make([]byte, meta.MSK_HEADER_SIZE+len(data))

	offset := 0
	copy(file[offset:], []byte(meta.MSK_MAGIC_VALUE))

	offset += meta.MSK_MAGIC_SIZE
	file[offset] = meta.MSK_FILE_VERSION

	offset += meta.MSK_VERSION_SIZE
	copy(file[offset:], salt)

	offset += meta.MSK_SALT_SIZE
	copy(file[offset:], nonce)

	offset += meta.MSK_NONCE_SIZE
	copy(file[offset:], data)

	return file, nil
}

func UnmarshalFile(data []byte) (salt, nonce, secret []byte, err error) {
	if len(data) < meta.MSK_HEADER_SIZE {
		return nil, nil, nil, ErrCorruptedFile
	}

	if string(data[:meta.MSK_MAGIC_SIZE]) != meta.MSK_MAGIC_VALUE {
		return nil, nil, nil, ErrCorruptedFile
	}

	if len(data) <= meta.MSK_MAGIC_SIZE {
		return nil, nil, nil, ErrCorruptedFile
	}

	if data[meta.MSK_MAGIC_SIZE] != meta.MSK_FILE_VERSION {
		return nil, nil, nil, ErrUnsupportedFileVersion
	}

	offset := meta.MSK_MAGIC_SIZE + meta.MSK_VERSION_SIZE

	salt = data[offset : offset+meta.MSK_SALT_SIZE]
	offset += meta.MSK_SALT_SIZE

	nonce = data[offset : offset+meta.MSK_NONCE_SIZE]
	offset += meta.MSK_NONCE_SIZE

	secret = data[offset:]

	return salt, nonce, secret, nil
}
