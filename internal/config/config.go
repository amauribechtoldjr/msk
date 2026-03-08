package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/amauribechtoldjr/msk/internal/domain"
	"github.com/amauribechtoldjr/msk/internal/files"
	"github.com/amauribechtoldjr/msk/internal/format"
	"github.com/amauribechtoldjr/msk/internal/vault"
	"github.com/amauribechtoldjr/msk/internal/wipe"
)

var (
	ErrConfigNotFound = errors.New("config file not found, run 'msk config' first")
	ErrInvalidConfig  = errors.New("master key verification failed")
)

const MSK_CONFIG_NAME = "msk-config"

type Config struct {
	Path string
}

func NewConfig() (*Config, error) {
	path, err := files.MSKConfigPath("config.msk")
	if err != nil {
		return &Config{}, err
	}

	return &Config{Path: path}, nil
}

func (c *Config) Load(vault vault.Vault) (string, error) {
	data, err := os.ReadFile(c.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrConfigNotFound
		}
		return "", err
	}

	salt, nonce, data, err := format.UnmarshalFile(data)
	if err != nil {
		return "", err
	}

	decryptedBytes, err := vault.Decrypt(salt, nonce, data)
	if err != nil {
		return "", ErrInvalidConfig
	}

	secret, err := format.UnmarshalSecret(decryptedBytes)
	if err != nil {
		return "", err
	}
	defer wipe.Bytes(secret.Password)

	if secret.Name != MSK_CONFIG_NAME {
		return "", ErrInvalidConfig
	}

	return string(secret.Password), nil
}

func (c *Config) Save(vault vault.Vault, vaultPath string) error {
	if err := os.MkdirAll(filepath.Dir(c.Path), 0o700); err != nil {
		return err
	}

	secret := domain.Secret{
		Name:     MSK_CONFIG_NAME,
		Password: []byte(vaultPath),
	}

	fileBytes := format.MarshalSecret(secret)

	saltedGCM, err := vault.Encrypt(fileBytes)
	if err != nil {
		return err
	}

	finalBytes, err := format.MarshalFile(saltedGCM.Salt, saltedGCM.Nonce, saltedGCM.CipherData)
	if err != nil {
		return err
	}

	// TODO: refactor to use SaveFile
	tmpPath := c.Path + ".tmp"
	if err := os.WriteFile(tmpPath, finalBytes, 0o600); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, c.Path); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}

func (c *Config) DefaultVaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".msk", "vault"), nil
}
