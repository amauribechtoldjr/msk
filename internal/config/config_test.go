package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/amauribechtoldjr/msk/internal/encryption"
)

func setupTestConfig(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	configPathOverride = filepath.Join(tmpDir, "config.msk")
	t.Cleanup(func() {
		configPathOverride = ""
	})
}

func TestSaveAndLoad(t *testing.T) {
	t.Run("should save and load vault path with correct key", func(t *testing.T) {
		setupTestConfig(t)

		enc := encryption.NewArgonCrypt()
		enc.ConfigMK([]byte("test-master-key"))

		vaultPath := "/home/user/.msk/vault"
		err := Save(enc, vaultPath)
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		loaded, err := Load(enc)
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
		setupTestConfig(t)

		enc := encryption.NewArgonCrypt()
		enc.ConfigMK([]byte("correct-key"))

		err := Save(enc, "/some/path")
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		enc.ConfigMK([]byte("wrong-key"))

		_, err = Load(enc)
		if !errors.Is(err, ErrInvalidConfig) {
			t.Fatalf("expected ErrInvalidConfig, got %v", err)
		}
	})
}

func TestLoadNotFound(t *testing.T) {
	t.Run("should return ErrConfigNotFound when file does not exist", func(t *testing.T) {
		setupTestConfig(t)

		enc := encryption.NewArgonCrypt()
		enc.ConfigMK([]byte("some-key"))

		_, err := Load(enc)
		if !errors.Is(err, ErrConfigNotFound) {
			t.Fatalf("expected ErrConfigNotFound, got %v", err)
		}
	})
}

func TestExists(t *testing.T) {
	t.Run("should return false when config does not exist", func(t *testing.T) {
		setupTestConfig(t)

		exists, err := Exists()
		if err != nil {
			t.Fatalf("Exists failed: %v", err)
		}

		if exists {
			t.Fatal("expected config to not exist")
		}
	})

	t.Run("should return true when config exists", func(t *testing.T) {
		setupTestConfig(t)

		enc := encryption.NewArgonCrypt()
		enc.ConfigMK([]byte("test-key"))

		err := Save(enc, "/some/path")
		if err != nil {
			t.Fatalf("Save failed: %v", err)
		}

		exists, err := Exists()
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
		path, err := DefaultVaultPath()
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
