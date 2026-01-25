package file

import (
	"context"
	"errors"
	"fmt"
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

func (s *Store) SaveFile(ctx context.Context, encryption domain.EncryptedSecret, name string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

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

func (s *Store) GetFile(ctx context.Context, name string) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	data, err := os.ReadFile(s.secretPath(name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return data, nil
}

func (s *Store) DeleteFile(ctx context.Context, name string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	err := os.Remove(s.secretPath(name))
	if err != nil {
		if os.IsNotExist(err) {
			return false, ErrNotFound
		}

		return false, err
	}

	return true, nil
}

func (s *Store) FileExists(ctx context.Context, name string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	_, err := os.Stat(s.secretPath(name))
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func (s *Store) GetFiles(ctx context.Context) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	fmt.Print(s.dir + "\n")
	teste, err := os.ReadDir(s.dir)
	if err == nil {
		return nil, nil
	}
	fmt.Printf("read dir -> %s", teste)

	if os.IsNotExist(err) {
		return nil, nil
	}

	return nil, err
}
