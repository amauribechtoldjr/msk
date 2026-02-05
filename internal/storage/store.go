package storage

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/amauribechtoldjr/msk/internal/domain"
)

var ErrNotFound = errors.New("secret not found")

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

	if _, err := tmpFile.Write([]byte("MSK")); err != nil {
		return err
	}

	if _, err := tmpFile.Write([]byte{1}); err != nil {
		return err
	}

	if _, err := tmpFile.Write(encryption.Salt[:]); err != nil {
		return err
	}

	if _, err := tmpFile.Write(encryption.Nonce[:]); err != nil {
		return err
	}

	if _, err := tmpFile.Write(encryption.Data); err != nil {
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
	err := os.Remove(s.secretPath(name))
	if err != nil {
		if os.IsNotExist(err) {
			return false, ErrNotFound
		}

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

	for _, fName := range files {
		info, err := fName.Info()
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			continue
		}

		names = append(names, fName.Name())
	}

	return names, err
}
