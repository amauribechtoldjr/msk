package vault

import (
	"errors"
	"reflect"
	"testing"

	"github.com/amauribechtoldjr/msk/internal/format"
	"github.com/amauribechtoldjr/msk/internal/meta"
)

func TestNewMSKVault(t *testing.T) {
	t.Run("should initialize the struct correctly", func(t *testing.T) {
		var crypt Vault = NewVault()
		_, ok := crypt.(*vault)

		if !ok {
			t.Fatal("expected and variable of type vault")
		}
	})
}

func newConfiguredCrypt(masterKey string) Vault {
	return NewVaultWithMK([]byte(masterKey))
}

func TestEncrypt(t *testing.T) {
	t.Run("should encrypt data successfully", func(t *testing.T) {
		crypt := newConfiguredCrypt("master-password")
		plaintext := []byte("s3cur3p@ss")

		result, err := crypt.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(result.CipherData) == 0 {
			t.Fatal("expected non-empty cipher data")
		}

		if len(result.Salt) != meta.MSK_SALT_SIZE {
			t.Fatal("expected valid salt")
		}

		if len(result.Nonce) == 0 {
			t.Fatal("expected valid nonce")
		}
	})

	t.Run("should produce different cipher data for same input", func(t *testing.T) {
		crypt := newConfiguredCrypt("master-password")
		plaintext := []byte("s3cur3p@ss")

		enc1, err := crypt.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("first encrypt failed: %v", err)
		}

		enc2, err := crypt.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("second encrypt failed: %v", err)
		}

		if reflect.DeepEqual(enc1.CipherData, enc2.CipherData) {
			t.Fatal("expected different cipher data due to random salt/nonce")
		}
	})

	t.Run("should return error when master key is not configured", func(t *testing.T) {
		crypt := NewVault()

		_, err := crypt.Encrypt([]byte("pass"))
		if err == nil {
			t.Fatal("expected error when master key is empty")
		}
	})
}

func TestDecrypt(t *testing.T) {
	t.Run("should decrypt encrypted data correctly (round-trip)", func(t *testing.T) {
		crypt := newConfiguredCrypt("master-password")
		plaintext := []byte("p@ssw0rd!")
		expected := make([]byte, len(plaintext))
		copy(expected, plaintext)

		encrypted, err := crypt.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		decrypted, err := crypt.Decrypt(encrypted.Salt, encrypted.Nonce, encrypted.CipherData)
		if err != nil {
			t.Fatalf("decrypt failed: %v", err)
		}

		if !reflect.DeepEqual(decrypted, expected) {
			t.Fatalf("expected %v, got %v", expected, decrypted)
		}
	})

	t.Run("should return ErrDecryption when master key is wrong", func(t *testing.T) {
		crypt := newConfiguredCrypt("correct-password")

		encrypted, err := crypt.Encrypt([]byte("pass"))
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		wrongCrypt := newConfiguredCrypt("wrong-password")
		_, err = wrongCrypt.Decrypt(encrypted.Salt, encrypted.Nonce, encrypted.CipherData)
		if err == nil {
			t.Fatal("expected error with wrong master key")
		}

		if !errors.Is(err, ErrDecryption) {
			t.Fatalf("expected ErrDecryption, got %v", err)
		}
	})

	t.Run("should return ErrDecryption when cipher data is tampered", func(t *testing.T) {
		crypt := newConfiguredCrypt("master-password")

		encrypted, err := crypt.Encrypt([]byte("pass"))
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		// Flip a byte in the cipher data portion
		encrypted.CipherData[0] ^= 0xFF

		_, err = crypt.Decrypt(encrypted.Salt, encrypted.Nonce, encrypted.CipherData)
		if err == nil {
			t.Fatal("expected error with tampered cipher data")
		}

		if !errors.Is(err, ErrDecryption) {
			t.Fatalf("expected ErrDecryption, got %v", err)
		}
	})

	t.Run("should return error when master key is empty", func(t *testing.T) {
		crypt := NewVault()
		salt, _ := format.RandomBytes(meta.MSK_SALT_SIZE)

		_, err := crypt.Decrypt(salt, []byte("nonce"), []byte("data"))
		if err == nil {
			t.Fatal("expected error when master key is empty")
		}
	})
}
