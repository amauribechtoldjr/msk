package file

import "testing"

// TODO: add tests for different OS path separators (or do them differently).
// Current tests assume Windows-style separators.

func TestSecretPath(t *testing.T) {
	store := &Store{dir: "/secrets"}
	expected := "\\secrets\\mysecret.msk"

	t.Run("should return correct secret path", func(t *testing.T) {
		result := store.secretPath("mysecret")

		if result != expected {
			t.Errorf("secretPath() = %v; want %v", result, expected)
		}
	})
}

func TestSecretPathLowerCase(t *testing.T) {
	store := &Store{dir: "/secrets"}
	expected := "\\secrets\\mysecret.msk"

	t.Run("should return correct secret path", func(t *testing.T) {
		result := store.secretPath("MYSECRET")

		if result != expected {
			t.Errorf("secretPath() = %v; want %v", result, expected)
		}
	})
}
