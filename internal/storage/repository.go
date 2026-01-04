package storage

import (
	"context"

	"github.com/amauribechtoldjr/msk/internal/domain"
)

type Repository interface {
	SaveFile(ctx context.Context, encryption domain.EncryptedSecret, name string) error
	GetFile(ctx context.Context, name string) ([]byte, error)
	FileExists(ctx context.Context, name string) (bool, error)
}