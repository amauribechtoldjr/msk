package vault

import (
	"crypto/sha256"
	"reflect"
	"testing"

	"github.com/amauribechtoldjr/msk/internal/format"
	"github.com/amauribechtoldjr/msk/internal/meta"
)

func TestGetSessionToken(t *testing.T) {
	t.Run("should return exactly 32 bytes", func(t *testing.T) {
		token, err := format.RandomBytes(meta.SESSION_TOKEN_SIZE)
		if err != nil {
			t.Fatal("failed to generate token")
		}

		key, err := DeriveSessionToken(token)
		if err != nil {
			t.Fatal("failed to derive session key")
		}

		if len(key) != 32 {
			t.Fatalf("expected key size: 32, got: %v", len(key))
		}
	})

	t.Run("should produce identical output for same token", func(t *testing.T) {
		token, err := format.RandomBytes(meta.SESSION_TOKEN_SIZE)
		if err != nil {
			t.Fatal("failed to generate token")
		}

		key1, err := DeriveSessionToken(token)
		if err != nil {
			t.Fatal("failed to derive session key")
		}

		key2, err := DeriveSessionToken(token)
		if err != nil {
			t.Fatal("failed to derive session key")
		}

		if !reflect.DeepEqual(key1, key2) {
			t.Fatal("expected identical output for same token")
		}
	})

	t.Run("should produce different output for different tokens", func(t *testing.T) {
		token1, err := format.RandomBytes(meta.SESSION_TOKEN_SIZE)
		if err != nil {
			t.Fatal("failed to generate token")
		}

		token2, err := format.RandomBytes(meta.SESSION_TOKEN_SIZE)
		if err != nil {
			t.Fatal("failed to generate token")
		}

		key1, err := DeriveSessionToken(token1)
		if err != nil {
			t.Fatal("failed to derive session key")
		}

		key2, err := DeriveSessionToken(token2)
		if err != nil {
			t.Fatal("failed to derive session key")
		}

		if reflect.DeepEqual(key1, key2) {
			t.Fatal("expected different output for different tokens")
		}
	})

	t.Run("should return error when empty token", func(t *testing.T) {
		_, err := DeriveSessionToken([]byte{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("should return error when nil token", func(t *testing.T) {
		_, err := DeriveSessionToken(nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("should match sha256 sum of token", func(t *testing.T) {
		token, err := format.RandomBytes(meta.SESSION_TOKEN_SIZE)
		if err != nil {
			t.Fatal("failed to generate token")
		}

		key, err := DeriveSessionToken(token)
		if err != nil {
			t.Fatal("failed to derive session key")
		}

		expected := sha256.Sum256(token)
		if !reflect.DeepEqual(key, expected[:]) {
			t.Fatal("key does not match expected sha256 hash of token")
		}
	})
}
