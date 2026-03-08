package storage

import (
	"errors"
	"os"
	"strings"

	"github.com/amauribechtoldjr/msk/internal/files"
)

var ErrNotFound = errors.New("secret not found")
var ErrInvalidSecret = errors.New("secret invalid")

type Repository interface {
	FileExists(name string) (bool, error)
	GetFile(name string) ([]byte, error)
	SaveFile(encryptedFile []byte, name string) error
	DeleteFile(name string) error
	GetFiles() ([]string, error)
}

type Store struct {
	Path string
}

func NewStore(path string) (*Store, error) {
	if err := os.MkdirAll(path, 0o700); err != nil {
		return nil, err
	}

	return &Store{path}, nil
}

func (s *Store) SaveFile(encryptedFile []byte, name string) error {
	return files.WriteAtomicFile(s.getFilePath(name), encryptedFile, 0o600)
}

func (s *Store) GetFile(name string) ([]byte, error) {
	return files.ReadFile(s.getFilePath(name), ErrNotFound)
}

func (s *Store) FileExists(name string) (bool, error) {
	return files.FileExists(s.getFilePath(name))
}

func (s *Store) DeleteFile(name string) error {
	filePath := s.getFilePath(name)

	exists, err := files.FileExists(filePath)
	if err != nil {
		return err
	}

	if !exists {
		return ErrNotFound
	}

	return os.Remove(filePath)
}

func (s *Store) GetFiles() ([]string, error) {
	files, err := os.ReadDir(s.Path)
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
