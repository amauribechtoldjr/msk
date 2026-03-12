package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/amauribechtoldjr/msk/internal/domain"
	"github.com/amauribechtoldjr/msk/internal/files"
	"github.com/amauribechtoldjr/msk/internal/format"
	"github.com/amauribechtoldjr/msk/internal/prompt"
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

func (c *Config) PromptVaultPath() (string, error) {
	defaultPath, err := c.DefaultVaultPath()
	if err != nil {
		return "", fmt.Errorf("failed to get default vault path: %w", err)
	}

	vaultPath, err :=
		prompt.ReadString(fmt.Sprintf("Enter vault path (press Enter for default: %s): ", defaultPath))

	if err != nil {
		return "", errors.New("invalid vault path")
	}

	vaultPath = strings.TrimSpace(vaultPath)
	if vaultPath == "" {
		vaultPath = defaultPath
	}

	return vaultPath, nil
}

func (c *Config) CheckOverwrite() (bool, error) {
	confirmed, err := prompt.ReadBoolean("Config already exists. Overwrite? (y/N): ")
	if err != nil {
		return false, err
	}

	return confirmed, nil
}

func (c *Config) CreateVault() (string, error) {
	exists, err := c.Exists()
	if err != nil {
		return "", err
	}

	var shouldOverwrite bool
	if exists {
		shouldOverwrite, err = c.CheckOverwrite()
		if err != nil {
			return "", err
		}

		if !shouldOverwrite {
			return "", errors.New("overwrite cancelled.")
		}
	}

	vaultPath, err := c.PromptVaultPath()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(vaultPath, 0o600); err != nil {
		return "", fmt.Errorf("failed to create vault directory: %w", err)
	}

	return vaultPath, nil
}

func (c *Config) Load(vault vault.Vault) (string, error) {
	data, err := files.ReadFile(c.Path, ErrConfigNotFound)
	if err != nil {
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

	return files.WriteAtomicFile(c.Path, finalBytes, 0o600)
}

func (c *Config) DefaultVaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".msk", "vault"), nil
}

func (c *Config) Exists() (bool, error) {
	exists, err := files.FileExists(c.Path)
	if err != nil {
		return false, err
	}

	return exists, nil
}
