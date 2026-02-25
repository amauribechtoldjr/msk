package storage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var ErrNotFound = errors.New("secret not found")
var ErrInvalidSecret = errors.New("secret invalid")

type Repository interface {
	FileExists(name string) bool
	GetFile(name string) ([]byte, error)
	SaveFile(encryptedFile []byte, name string) error
	DeleteFile(name string) error
	GetFiles() ([]string, error)
}

type Store struct {
	dir string
}

func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}

	return &Store{dir: dir}, nil
}

func (s *Store) SaveFile(encryptedFile []byte, name string) error {
	path := s.getFilePath(name)
	tmpPath := path + ".tmp"

	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}

	defer func() {
		tmpFile.Close()
		os.Remove(tmpPath)
	}()

	if _, err := tmpFile.Write(encryptedFile); err != nil {
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

	return nil
}

func (s *Store) GetFile(name string) ([]byte, error) {
	data, err := os.ReadFile(s.getFilePath(name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return data, nil
}

func (s *Store) DeleteFile(name string) error {
	getFilePath := s.getFilePath(name)

	info, err := os.Stat(getFilePath)
	if err != nil {
		return ErrNotFound
	}

	if info.IsDir() {
		return ErrInvalidSecret
	}

	err = os.Remove(getFilePath)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) FileExists(name string) bool {
	_, err := os.Stat(s.getFilePath(name))
	if err == nil {
		return true
	}

	return false
}

func (s *Store) GetFiles() ([]string, error) {
	files, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
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
