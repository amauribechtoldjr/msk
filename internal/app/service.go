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
	ErrSecretExists = errors.New("secret already exists")
	ErrSecretNotFound = errors.New("secret not found")
)

type MSKService interface {
	AddSecret(ctx context.Context, name, rawS string) error
	GetSecret(ctx context.Context, name string) error
	ListAll(ctx context.Context) error
	ConfigMK(ctx context.Context, mk []byte)
}

type Service struct {
	repo storage.Repository
	crypto encryption.Encryption
}

func NewMSKService(r storage.Repository, c encryption.Encryption) *Service {
	return &Service{
		crypto: c,
		repo: r,
	}
}

func (s *Service) ConfigMK(ctx context.Context, mk []byte) {
	s.crypto.ConfigMK(mk) 
}

func (s *Service) AddSecret(ctx context.Context, name, rawS string) error {
	if name == "" {
		return errors.New("secret name cannot be empty")
	}

	exists, err := s.repo.FileExists(ctx, name)
	if err != nil {
		return err
	}

	if exists {
		return ErrSecretExists
	}

	secret := domain.Secret{
		Name: name,
		Password: []byte(rawS),
		CreatedAt: time.Now().UTC(),
	}

	//TODO: continue refactor from here (save secret from main.go)
	encryptionResult, err := s.crypto.Encrypt(secret)
	if err != nil {
		return err
	}

	return s.repo.SaveFile(ctx, encryptionResult, name)
}

func (s *Service) GetSecret(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("secret name cannot be empty")
	}

	exists, err := s.repo.FileExists(ctx, name)
	if err != nil {
		return err
	}

	if !exists {
		return ErrSecretNotFound
	}

	fileData, err := s.repo.GetFile(ctx, name)
	if err != nil {
		return err
	}

	secretData, err := s.crypto.Decrypt(fileData)
	if err != nil {
		return err
	}

	fmt.Println(string(secretData.Password))

	return nil
}

func (s *Service) ListAll(ctx context.Context) error {
	return nil
}