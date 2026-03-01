package session

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/amauribechtoldjr/msk/internal/vault"
	"github.com/amauribechtoldjr/msk/internal/wipe"
)

var (
	ErrSessionExpired  = errors.New("session expired")
	ErrSessionInvalid  = errors.New("session file invalid or corrupted")
	ErrSessionNotFound = errors.New("session file not found")
)

var sessionPathOverride string

type Session struct {
	path string
}

func New() (*Session, error) {
	if sessionPathOverride != "" {
		return &Session{path: sessionPathOverride}, nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return &Session{path: filepath.Join(configDir, "msk", "session")}, nil
}

func (s *Session) Create(v vault.MSKVault) (string, error) {
	tokenBytes, err := randomBytes(vault.SESSION_TOKEN_SIZE)
	if err != nil {
		return "", err
	}
	defer wipe.Bytes(tokenBytes)

	token := hex.EncodeToString(tokenBytes)

	nonce, cipherMkData, err := v.CreateSession()
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

func (s *Session) Load(token string) ([]byte, error) {
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

	sessionKey := deriveKey(tokenBytes)
	defer wipe.Bytes(sessionKey)

	nonce := data[0:vault.SESSION_NONCE_SIZE]
	ciphertext := data[vault.SESSION_HEADER_SIZE:]
	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return nil, ErrSessionInvalid
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrSessionInvalid
	}

	mk, err := gcm.Open(nil, nonce, ciphertext, nil)
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

func (s *Session) IsActive(token string) bool {
	_, err := s.Load(token)
	return err == nil
}

func deriveKey(tokenBytes []byte) []byte {
	sum := sha256.Sum256(tokenBytes)
	key := make([]byte, 32)
	copy(key, sum[:])
	return key
}

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
