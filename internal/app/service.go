package app

import (
	"errors"

	"github.com/amauribechtoldjr/msk/internal/domain"
	"github.com/amauribechtoldjr/msk/internal/storage"
	"github.com/amauribechtoldjr/msk/internal/vault"
	"github.com/amauribechtoldjr/msk/internal/wipe"
)

var (
	ErrSecretExists   = errors.New("secret already exists")
	ErrSecretNotFound = errors.New("secret not found")
)

type MSKService struct {
	repo   storage.Repository
	crypto vault.Vault
}

func NewMSKService(r storage.Repository, c vault.Vault) *MSKService {
	return &MSKService{
		crypto: c,
		repo:   r,
	}
}

func (s *MSKService) ConfigMK(mk []byte) {
	s.crypto.ConfigMK(mk)
}

func (s *MSKService) DestroyMK() {
	s.crypto.DestroyMK()
}

func (s *MSKService) DeleteSecret(name string) error {
	if !s.repo.FileExists(name) {
		return ErrSecretNotFound
	}

	return s.repo.DeleteFile(name)
}

func (s *MSKService) AddSecret(name string, rawP []byte) error {
	if s.repo.FileExists(name) {
		return ErrSecretExists
	}

	secret := domain.Secret{
		Name:     name,
		Password: rawP,
	}
	defer wipe.Bytes(secret.Password)

	encryptionResult, err := s.crypto.EncryptSecret(secret)
	if err != nil {
		return err
	}

	return s.repo.SaveFile(encryptionResult, name)
}

func (s *MSKService) UpdateSecret(name string, rawP []byte) error {
	if !s.repo.FileExists(name) {
		return ErrSecretNotFound
	}

	secret := domain.Secret{
		Name:     name,
		Password: rawP,
	}
	defer wipe.Bytes(secret.Password)

	encryptionResult, err := s.crypto.EncryptSecret(secret)
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

	secretData, err := s.crypto.DecryptSecret(fileData)
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
