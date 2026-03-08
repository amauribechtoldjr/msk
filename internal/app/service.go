package app

import (
	"errors"

	"github.com/amauribechtoldjr/msk/internal/domain"
	"github.com/amauribechtoldjr/msk/internal/format"
	"github.com/amauribechtoldjr/msk/internal/storage"
	"github.com/amauribechtoldjr/msk/internal/vault"
	"github.com/amauribechtoldjr/msk/internal/wipe"
)

var (
	ErrSecretExists   = errors.New("secret already exists")
	ErrSecretNotFound = errors.New("secret not found")
)

type Service interface {
	DeleteSecret(name string) error
	AddSecret(name string, rawP []byte) error
	UpdateSecret(name string, rawP []byte) error
	GetSecret(name string) ([]byte, error)
	ListSecrets() ([]string, error)
}

type MSKService struct {
	repo  storage.Repository
	vault vault.Vault
}

func NewMSKService(r storage.Repository, v vault.Vault) Service {
	return &MSKService{
		vault: v,
		repo:  r,
	}
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

	secretBytes := format.MarshalSecret(secret)

	saltedGCM, err := s.vault.Encrypt(secretBytes)
	if err != nil {
		return err
	}

	fileBytes, err := format.MarshalFile(saltedGCM.Salt, saltedGCM.Nonce, saltedGCM.CipherData)
	if err != nil {
		return err
	}

	return s.repo.SaveFile(fileBytes, name)
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

	secretBytes := format.MarshalSecret(secret)

	saltedGCM, err := s.vault.Encrypt(secretBytes)
	if err != nil {
		return err
	}

	fileBytes, err := format.MarshalFile(saltedGCM.Salt, saltedGCM.Nonce, saltedGCM.CipherData)
	if err != nil {
		return err
	}

	return s.repo.SaveFile(fileBytes, name)
}

func (s *MSKService) GetSecret(name string) ([]byte, error) {
	if !s.repo.FileExists(name) {
		return nil, ErrSecretNotFound
	}

	fileData, err := s.repo.GetFile(name)
	if err != nil {
		return nil, err
	}

	salt, nonce, data, err := format.UnmarshalFile(fileData)
	if err != nil {
		return nil, err
	}

	decryptedBytes, err := s.vault.Decrypt(salt, nonce, data)
	if err != nil {
		return nil, err
	}
	defer wipe.Bytes(decryptedBytes)

	secret, err := format.UnmarshalSecret(decryptedBytes)
	if err != nil {
		return nil, err
	}

	return secret.Password, nil
}

func (s *MSKService) ListSecrets() ([]string, error) {
	files, err := s.repo.GetFiles()
	if err != nil {
		return nil, err
	}

	return files, nil
}
