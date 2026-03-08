package session

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/amauribechtoldjr/msk/internal/files"
	"github.com/amauribechtoldjr/msk/internal/format"
	"github.com/amauribechtoldjr/msk/internal/gcm"
	"github.com/amauribechtoldjr/msk/internal/meta"
	"github.com/amauribechtoldjr/msk/internal/wipe"
)

var (
	ErrSessionExpired  = errors.New("session expired")
	ErrSessionInvalid  = errors.New("session file invalid or corrupted")
	ErrSessionNotFound = errors.New("session file not found")
)

type Session interface {
	LoadFile(token string) (*BinarySession, error)
	StoreSession(sealedGCM *gcm.SealedCGM) error
	GetSessionToken() ([]byte, error)
	Destroy() error
}

type session struct {
	path string
}

type BinarySession struct {
	Token      []byte
	Nonce      []byte
	CipherData []byte
}

func New() (Session, error) {
	path, err := files.MSKConfigPath("session.msk")
	if err != nil {
		return nil, err
	}

	return &session{path: path}, nil
}

func (s *session) LoadFile(token string) (*BinarySession, error) {
	data, err := files.ReadFile(s.path, ErrSessionNotFound)
	if err != nil {
		return nil, err
	}

	if len(data) < meta.SESSION_HEADER_SIZE+1 {
		return nil, ErrSessionInvalid
	}

	expiry := int64(binary.BigEndian.Uint64(data[meta.SESSION_NONCE_SIZE:meta.SESSION_HEADER_SIZE]))
	if time.Now().Unix() > expiry {
		return nil, ErrSessionExpired
	}

	tokenBytes, err := hex.DecodeString(token)
	if err != nil || len(tokenBytes) != meta.SESSION_TOKEN_SIZE {
		return nil, ErrSessionInvalid
	}

	nonce := data[0:meta.SESSION_NONCE_SIZE]
	cipherData := data[meta.SESSION_HEADER_SIZE:]

	return &BinarySession{
		Token:      tokenBytes,
		Nonce:      nonce,
		CipherData: cipherData,
	}, nil
}

func (s *session) GetSessionToken() ([]byte, error) {
	tokenBytes, err := format.RandomBytes(meta.SESSION_TOKEN_SIZE)
	if err != nil {
		return nil, err
	}

	return tokenBytes, nil
}

func (s *session) StoreSession(sealedGCM *gcm.SealedCGM) error {
	expiry := time.Now().Add(meta.SESSION_TTL).Unix()
	file := make([]byte, meta.SESSION_HEADER_SIZE+len(sealedGCM.CipherData))
	defer wipe.Bytes(file)
	copy(file[0:meta.SESSION_NONCE_SIZE], sealedGCM.Nonce)
	binary.BigEndian.PutUint64(file[meta.SESSION_NONCE_SIZE:meta.SESSION_HEADER_SIZE], uint64(expiry))
	copy(file[meta.SESSION_HEADER_SIZE:], sealedGCM.CipherData)

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}

	return files.WriteFile(s.path, file, 0o600)
}

func (s *session) Refresh() error {
	data, err := files.ReadFile(s.path, ErrSessionNotFound)
	if err != nil {
		return err
	}
	defer wipe.Bytes(data)

	if len(data) < meta.SESSION_HEADER_SIZE+1 {
		return ErrSessionInvalid
	}

	newExpiry := time.Now().Add(meta.SESSION_TTL).Unix()
	binary.BigEndian.PutUint64(data[meta.SESSION_NONCE_SIZE:meta.SESSION_HEADER_SIZE], uint64(newExpiry))

	return files.WriteFile(s.path, data, 0o600)
}

func (s *session) Destroy() error {
	err := os.Remove(s.path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
