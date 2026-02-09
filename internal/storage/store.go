package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/amauribechtoldjr/msk/internal/domain"
)

var ErrNotFound = errors.New("secret not found")
var ErrInvalidSecret = errors.New("secret invalid")

type Store struct {
	dir string
}

func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}

	return &Store{dir: dir}, nil
}

func (s *Store) SaveFile(encryption domain.EncryptedSecret, name string) error {
	path := s.secretPath(name)
	tmpPath := path + ".tmp"

	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}

	defer func() {
		tmpFile.Close()
		os.Remove(tmpPath)
	}()

	msk := []byte("MSK")
	version := []byte{1}
	content := []byte{}

	content = append(content, msk...)
	content = append(content, version...)
	content = append(content, encryption.Salt[:]...)
	content = append(content, encryption.Nonce[:]...)
	content = append(content, encryption.Data...)

	if _, err := tmpFile.Write(content); err != nil {
		return err
	}

	if err := tmpFile.Sync(); err != nil {
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}

	dir, err := os.Open(filepath.Dir(path))
	if err == nil {
		defer dir.Close()
		_ = dir.Sync()
	}

	// TODO: study atomically writing strategies

	return nil
}

func (s *Store) GetFile(name string) ([]byte, error) {
	data, err := os.ReadFile(s.secretPath(name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return data, nil
}

func (s *Store) DeleteFile(name string) (bool, error) {
	secretPath := s.secretPath(name)
	fmt.Printf("secretPath: %v \n", secretPath)
	info, err := os.Stat(secretPath)
	if err != nil {
		return false, ErrNotFound
	}

	if info.IsDir() {
		return false, ErrInvalidSecret
	}

	err = os.Remove(secretPath)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *Store) FileExists(name string) bool {
	_, err := os.Stat(s.secretPath(name))
	if err == nil {
		return true
	}

	return false
}

func (s *Store) GetFiles() ([]string, error) {
	files, err := os.ReadDir(s.dir)

	if err != nil {
		return nil, nil
	}

	var names []string

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			continue
		}

		if !strings.Contains(file.Name(), ".msk") {
			continue
		}

		names = append(names, file.Name())
	}

	return names, err
}
