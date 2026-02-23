package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/amauribechtoldjr/msk/internal/domain"
	"github.com/amauribechtoldjr/msk/internal/encryption"
	"github.com/amauribechtoldjr/msk/internal/wipe"
)

var (
	ErrConfigNotFound = errors.New("config file not found, run 'msk config' first")
	ErrInvalidConfig  = errors.New("master key verification failed")
)

const MSK_CONFIG_NAME = "msk-config"

var configPathOverride string

func Path() (string, error) {
	if configPathOverride != "" {
		return configPathOverride, nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "msk", "config.msk"), nil
}

func Exists() (bool, error) {
	path, err := Path()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func Load(enc encryption.Encryption) (string, error) {
	path, err := Path()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrConfigNotFound
		}
		return "", err
	}

	secret, err := enc.Decrypt(data)
	if err != nil {
		return "", ErrInvalidConfig
	}
	defer wipe.Bytes(secret.Password)

	if secret.Name != MSK_CONFIG_NAME {
		return "", ErrInvalidConfig
	}

	return string(secret.Password), nil
}

func Save(enc encryption.Encryption, vaultPath string) error {
	path, err := Path()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	secret := domain.Secret{
		Name:     MSK_CONFIG_NAME,
		Password: []byte(vaultPath),
	}

	encrypted, err := enc.Encrypt(secret)
	if err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, encrypted, 0o600); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}

func DefaultVaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".msk", "vault"), nil
}
