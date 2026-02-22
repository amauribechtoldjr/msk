package encryption

import (
	"errors"
	"reflect"
	"testing"

	"github.com/amauribechtoldjr/msk/internal/format"
)

func TestGetArgonDeriveKey(t *testing.T) {
	t.Run("should return exactly 32 bytes for every password/salt combination", func(t *testing.T) {
		masterPassword := []byte("master-pass")
		salt, err := randomBytes(format.MSK_SALT_SIZE)
		if err != nil {
			t.Fatal("failed to generate salt array")
		}

		expectedSize := 32
		key, err := getArgonDeriveKey(masterPassword, salt)
		if err != nil {
			t.Fatal("failed to generate argon derived key")
		}

		if len(key) != expectedSize {
			t.Fatalf("expected key size: %v, got: %v", expectedSize, len(key))
		}
	})

	t.Run("should produces identical output for same master and salt", func(t *testing.T) {
		masterPassword := []byte("master-pass")
		salt, err := randomBytes(format.MSK_SALT_SIZE)
		if err != nil {
			t.Fatal("failed to generate salt array")
		}

		key, err := getArgonDeriveKey(masterPassword, salt)
		if err != nil {
			t.Fatal("failed to generate argon derived key")
		}

		key2, err := getArgonDeriveKey(masterPassword, salt)
		if err != nil {
			t.Fatal("failed to generate argon derived key")
		}

		if !reflect.DeepEqual(key, key2) {
			t.Fatalf("failed to produce same output for equal entries")
		}
	})

	t.Run("should produces different output when different master pass", func(t *testing.T) {
		masterPassword := []byte("master-pass")
		salt, err := randomBytes(format.MSK_SALT_SIZE)
		if err != nil {
			t.Fatal("failed to generate salt array")
		}

		key, err := getArgonDeriveKey(masterPassword, salt)
		if err != nil {
			t.Fatal("failed to generate argon derived key")
		}

		masterPassword2 := []byte("master-pass2")

		key2, err := getArgonDeriveKey(masterPassword2, salt)
		if err != nil {
			t.Fatal("failed to generate argon derived key")
		}

		if reflect.DeepEqual(key, key2) {
			t.Fatal("failed to generate different outputs for different master pass")
		}
	})

	t.Run("should produces different output when different salt", func(t *testing.T) {
		masterPassword := []byte("master-pass")
		salt, err := randomBytes(format.MSK_SALT_SIZE)
		if err != nil {
			t.Fatal("failed to generate salt array")
		}

		key, err := getArgonDeriveKey(masterPassword, salt)
		if err != nil {
			t.Fatal("failed to generate argon derived key")
		}

		salt2, err := randomBytes(format.MSK_SALT_SIZE)
		if err != nil {
			t.Fatal("failed to generate salt array")
		}

		key2, err := getArgonDeriveKey(masterPassword, salt2)
		if err != nil {
			t.Fatal("failed to generate argon derived key")
		}

		if reflect.DeepEqual(key, key2) {
			t.Fatal("failed to generate different outputs for different master pass")
		}
	})

	t.Run("should return error when empty pass", func(t *testing.T) {
		masterPassword := []byte("")
		salt, err := randomBytes(format.MSK_SALT_SIZE)
		if err != nil {
			t.Fatal("failed to generate salt array")
		}

		_, err = getArgonDeriveKey(masterPassword, salt)
		if err == nil {
			t.Fatal("expected ErrInvalidPass, got no error")
		}

		if !errors.Is(err, ErrInvalidPass) {
			t.Fatalf("expected ErrInvalidPass, got %v", err)
		}
	})

	t.Run("should return error when empty salt", func(t *testing.T) {
		masterPassword := []byte("master-pass")
		salt := []byte{}

		_, err := getArgonDeriveKey(masterPassword, salt)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrInvalidSalt) {
			t.Fatalf("expected ErrInvalidPass, got %v", err)
		}
	})
}
