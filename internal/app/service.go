package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/amauribechtoldjr/msk/internal/domain"
	"github.com/amauribechtoldjr/msk/internal/encryption"
	"github.com/amauribechtoldjr/msk/internal/storage"
)

var (
	ErrSecretExists   = errors.New("secret already exists")
	ErrSecretNotFound = errors.New("secret not found")
)

type MSKService interface {
	AddSecret(ctx context.Context, name string, password []byte) error
	GetSecret(ctx context.Context, name string) ([]byte, error)
	DeleteSecret(ctx context.Context, name string) error
	ListSecrets(ctx context.Context) ([]string, error)
	ConfigMK(ctx context.Context, mk []byte)
}

type Service struct {
	repo   storage.Repository
	crypto encryption.Encryption
}

func NewMSKService(r storage.Repository, c encryption.Encryption) *Service {
	return &Service{
		crypto: c,
		repo:   r,
	}
}

func (s *Service) ConfigMK(ctx context.Context, mk []byte) {
	s.crypto.ConfigMK(mk)
}

func (s *Service) DeleteSecret(ctx context.Context, name string) error {
	_, err := s.repo.DeleteFile(ctx, name)

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) AddSecret(ctx context.Context, name string, rawP []byte) error {
	exists, err := s.repo.FileExists(ctx, name)
	if err != nil {
		return err
	}

	if exists {
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

	return s.repo.SaveFile(ctx, encryptionResult, name)
}

func (s *Service) GetSecret(ctx context.Context, name string) ([]byte, error) {
	exists, err := s.repo.FileExists(ctx, name)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrSecretNotFound
	}

	fileData, err := s.repo.GetFile(ctx, name)
	if err != nil {
		return nil, err
	}

	secretData, err := s.crypto.Decrypt(fileData)
	if err != nil {
		return nil, err
	}

	return secretData.Password, nil
}

func (s *Service) ListSecrets(ctx context.Context) ([]string, error) {
	files, err := s.repo.GetFiles(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Printf("files %s", files)
	return files, nil
}
