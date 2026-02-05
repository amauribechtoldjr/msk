package storage

import (
	"path/filepath"
	"testing"
)

// add tests for:
// wrong path separators on different OS
// case insensitivity

func TestSecretPath(t *testing.T) {
	store := &Store{dir: "/secrets"}
	expected := filepath.ToSlash("\\secrets\\mysecret.msk")

	t.Run("should return correct secret path", func(t *testing.T) {
		result := store.secretPath("mysecret")

		if result != expected {
			t.Errorf("secretPath() = %v; want %v", result, expected)
		}
	})
}

func TestSecretPathLowerCase(t *testing.T) {
	store := &Store{dir: "/secrets"}
	expected := filepath.ToSlash("\\secrets\\mysecret.msk")

	t.Run("should return correct secret path", func(t *testing.T) {
		result := store.secretPath("MYSECRET")

		if result != expected {
			t.Errorf("secretPath() = %v; want %v", result, expected)
		}
	})
}
