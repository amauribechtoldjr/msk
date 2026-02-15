package app

import (
	"errors"
	"time"

	"github.com/amauribechtoldjr/msk/internal/domain"
	"github.com/amauribechtoldjr/msk/internal/encryption"
	"github.com/amauribechtoldjr/msk/internal/storage"
)

var (
	ErrSecretExists   = errors.New("secret already exists")
	ErrSecretNotFound = errors.New("secret not found")
)

type MSKService struct {
	repo   storage.Repository
	crypto encryption.Encryption
}

func NewMSKService(r storage.Repository, c encryption.Encryption) *MSKService {
	return &MSKService{
		crypto: c,
		repo:   r,
	}
}

func (s *MSKService) ConfigMK(mk []byte) {
	s.crypto.ConfigMK(mk)
}

func (s *MSKService) DeleteSecret(name string) error {
	if !s.repo.FileExists(name) {
		return ErrSecretNotFound
	}

	err := s.repo.DeleteFile(name)
	if err != nil {
		return err
	}

	return nil
}

func (s *MSKService) AddSecret(name string, rawP []byte) error {
	if s.repo.FileExists(name) {
		return ErrSecretExists
	}

	secret := domain.Secret{
		Name:      name,
		Password:  rawP,
		CreatedAt: time.Now().UTC(),
	}

	encryptionResult, err := s.crypto.Encrypt(secret)
	if err != nil {
		return err
	}

	return s.repo.SaveFile(encryptionResult, name)
}

func (s *MSKService) GetSecret(name string) ([]byte, error) {
	if !s.repo.FileExists(name) {
		return nil, ErrSecretNotFound
	}

	fileData, err := s.repo.GetFile(name)
	if err != nil {
		return nil, err
	}

	secretData, err := s.crypto.Decrypt(fileData)
	if err != nil {
		return nil, err
	}

	return secretData.Password, nil
}

func (s *MSKService) ListSecrets() ([]string, error) {
	files, err := s.repo.GetFiles()
	if err != nil {
		return nil, err
	}

	return files, nil
}
