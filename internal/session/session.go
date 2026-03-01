package session

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/amauribechtoldjr/msk/internal/format"
	"github.com/amauribechtoldjr/msk/internal/vault"
	"github.com/amauribechtoldjr/msk/internal/wipe"
)

var (
	ErrSessionExpired  = errors.New("session expired")
	ErrSessionInvalid  = errors.New("session file invalid or corrupted")
	ErrSessionNotFound = errors.New("session file not found")
)

type Session struct {
	path string
}

func New() (*Session, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	return &Session{
		path: filepath.Join(configDir, "msk", "session.msk"),
	}, nil
}

func (s *Session) Create(v vault.Vault) (string, error) {
	tokenBytes, err := format.RandomBytes(vault.SESSION_TOKEN_SIZE)
	if err != nil {
		return "", err
	}
	defer wipe.Bytes(tokenBytes)

	token := hex.EncodeToString(tokenBytes)

	nonce, cipherMkData, err := v.CreateSession(tokenBytes)
	if err != nil {
		return "", err
	}

	expiry := time.Now().Add(vault.SESSION_TTL).Unix()
	file := make([]byte, vault.SESSION_HEADER_SIZE+len(cipherMkData))
	defer wipe.Bytes(file)
	copy(file[0:vault.SESSION_NONCE_SIZE], nonce)
	binary.BigEndian.PutUint64(file[vault.SESSION_NONCE_SIZE:vault.SESSION_HEADER_SIZE], uint64(expiry))
	copy(file[vault.SESSION_HEADER_SIZE:], cipherMkData)

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return "", err
	}

	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, file, 0o600); err != nil {
		return "", err
	}

	if err := os.Rename(tmpPath, s.path); err != nil {
		os.Remove(tmpPath)
		return "", err
	}
	return token, nil
}

func (s *Session) Load(token string, v vault.Vault) ([]byte, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrSessionNotFound
		}
		return nil, ErrSessionInvalid
	}

	if len(data) < vault.SESSION_HEADER_SIZE+1 {
		return nil, ErrSessionInvalid
	}

	expiry := int64(binary.BigEndian.Uint64(data[vault.SESSION_NONCE_SIZE:vault.SESSION_HEADER_SIZE]))
	if time.Now().Unix() > expiry {
		return nil, ErrSessionExpired
	}

	tokenBytes, err := hex.DecodeString(token)
	if err != nil || len(tokenBytes) != vault.SESSION_TOKEN_SIZE {
		return nil, ErrSessionInvalid
	}
	defer wipe.Bytes(tokenBytes)

	nonce := data[0:vault.SESSION_NONCE_SIZE]
	ciphertext := data[vault.SESSION_HEADER_SIZE:]

	mk, err := v.LoadSession(tokenBytes, nonce, ciphertext)
	if err != nil {
		return nil, ErrSessionInvalid
	}

	return mk, nil
}

func (s *Session) Refresh() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrSessionNotFound
		}
		return err
	}
	defer wipe.Bytes(data)

	if len(data) < vault.SESSION_HEADER_SIZE+1 {
		return ErrSessionInvalid
	}

	newExpiry := time.Now().Add(vault.SESSION_TTL).Unix()
	binary.BigEndian.PutUint64(data[vault.SESSION_NONCE_SIZE:vault.SESSION_HEADER_SIZE], uint64(newExpiry))

	tmpPath := s.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return nil
}

func (s *Session) Destroy() error {
	err := os.Remove(s.path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
