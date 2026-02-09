package storage

import (
	"path/filepath"
	"strings"
	"testing"
)

// add tests for:
// wrong path separators on different OS
// case insensitivity

var FILE_EXT = "msk"

func TestSecretPath(t *testing.T) {
	secretName := "mysecret"
	store := &Store{dir: t.TempDir()}
	expected := filepath.ToSlash(
		filepath.Join(
			store.dir,
			strings.Join([]string{secretName, FILE_EXT}, "."),
		),
	)

	t.Run("should return correct secret path", func(t *testing.T) {
		result := store.secretPath(secretName)

		if result != expected {
			t.Errorf("secretPath() = %v; want %v", result, expected)
		}
	})
}

func TestSecretPathLowerCase(t *testing.T) {
	secretName := "MYSECRET"
	store := &Store{dir: t.TempDir()}
	expected := filepath.ToSlash(
		filepath.Join(
			store.dir,
			strings.Join([]string{strings.ToLower(secretName), FILE_EXT}, "."),
		),
	)

	t.Run("should return correct secret path", func(t *testing.T) {
		result := store.secretPath(secretName)

		if result != expected {
			t.Errorf("secretPath() = %v; want %v", result, expected)
		}
	})
}
