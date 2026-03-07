package storage

import (
	"path/filepath"
	"strings"
	"testing"
)

var FILE_EXT = "msk"

func TestGetFilePath(t *testing.T) {
	secretName := "mysecret"
	store := &Store{Path: t.TempDir()}
	expected := filepath.Join(
		store.Path,
		strings.Join([]string{secretName, FILE_EXT}, "."),
	)

	t.Run("should return correct secret path", func(t *testing.T) {
		result := store.getFilePath(secretName)

		if result != expected {
			t.Errorf("getFilePath() = %v; want %v", result, expected)
		}
	})
}

func TestGetFilePathLowerCase(t *testing.T) {
	secretName := "MYSECRET"
	store := &Store{Path: t.TempDir()}
	expected := filepath.Join(
		store.Path,
		strings.Join([]string{strings.ToLower(secretName), FILE_EXT}, "."),
	)

	t.Run("should return correct secret path", func(t *testing.T) {
		result := store.getFilePath(secretName)

		if result != expected {
			t.Errorf("getFilePath() = %v; want %v", result, expected)
		}
	})
}
