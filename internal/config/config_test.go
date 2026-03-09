package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/amauribechtoldjr/msk/internal/files"
	"github.com/amauribechtoldjr/msk/internal/vault"
)

func newTestConfig(t *testing.T) *Config {
	t.Helper()

	tmpDir := t.TempDir()
	t.Setenv("AppData", tmpDir)           // windows
	t.Setenv("XDG_CONFIG_HOME", tmpDir)   // linux
	t.Setenv("HOME", tmpDir)              // macos

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig failed: %v", err)
	}
	return cfg
}

func TestSaveAndLoad(t *testing.T) {
	t.Run("should save and load vault path with correct key", func(t *testing.T) {
		cfg := newTestConfig(t)

		vault := vault.NewMSKVault()
		vault.ConfigMK([]byte("test-master-key"))

		vaultPath := "/home/user/.msk/vault"
		err := cfg.Save(vault, vaultPath)
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		loaded, err := cfg.Load(vault)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if loaded != vaultPath {
			t.Fatalf("expected vault path %q, got %q", vaultPath, loaded)
		}
	})
}

func TestLoadWrongKey(t *testing.T) {
	t.Run("should return ErrInvalidConfig with wrong key", func(t *testing.T) {
		cfg := newTestConfig(t)

		vault := vault.NewMSKVault()
		vault.ConfigMK([]byte("correct-key"))

		err := cfg.Save(vault, "/some/path")
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		vault.ConfigMK([]byte("wrong-key"))

		_, err = cfg.Load(vault)
		if !errors.Is(err, ErrInvalidConfig) {
			t.Fatalf("expected ErrInvalidConfig, got %v", err)
		}
	})
}

func TestLoadNotFound(t *testing.T) {
	t.Run("should return ErrConfigNotFound when file does not exist", func(t *testing.T) {
		cfg := newTestConfig(t)

		vault := vault.NewMSKVault()
		vault.ConfigMK([]byte("some-key"))

		_, err := cfg.Load(vault)
		if !errors.Is(err, ErrConfigNotFound) {
			t.Fatalf("expected ErrConfigNotFound, got %v", err)
		}
	})
}

func TestExists(t *testing.T) {
	t.Run("should return false when config does not exist", func(t *testing.T) {
		cfg := newTestConfig(t)

		exists, err := files.FileExists(cfg.Path)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}

		if exists {
			t.Fatal("expected config to not exist")
		}
	})

	t.Run("should return true when config exists", func(t *testing.T) {
		cfg := newTestConfig(t)

		vault := vault.NewMSKVault()
		vault.ConfigMK([]byte("test-key"))

		err := cfg.Save(vault, "/some/path")
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		exists, err := files.FileExists(cfg.Path)
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}

		if !exists {
			t.Fatal("expected config to exist")
		}
	})
}

func TestDefaultVaultPath(t *testing.T) {
	t.Run("should return a path under home directory", func(t *testing.T) {
		cfg := newTestConfig(t)

		path, err := cfg.DefaultVaultPath()
		if err != nil {
			t.Fatalf("DefaultVaultPath failed: %v", err)
		}

		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".msk", "vault")
		if path != expected {
			t.Fatalf("expected %q, got %q", expected, path)
		}
	})
}
