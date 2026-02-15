package encryption

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/amauribechtoldjr/msk/internal/domain"
)

func buildCipherData(salt [MSK_SALT_SIZE]byte, nonce [MSK_NONCE_SIZE]byte, cipherText []byte) []byte {
	buf := make([]byte, 0, MSK_HEADER_SIZE+len(cipherText))
	buf = append(buf, []byte(MSK_MAGIC_VALUE)...)
	buf = append(buf, MSK_FILE_VERSION)
	buf = append(buf, salt[:]...)
	buf = append(buf, nonce[:]...)
	buf = append(buf, cipherText...)
	return buf
}

func TestNewArgonCrypt(t *testing.T) {
	t.Run("should initialize the struct correctly", func(t *testing.T) {
		var crypt Encryption = NewArgonCrypt()
		_, ok := crypt.(*ArgonCrypt)

		if !ok {
			t.Fatal("expected and variable of type ArgonCrypt")
		}
	})
}

func TestConfigMk(t *testing.T) {
	t.Run("should set the master key correctly", func(t *testing.T) {
		crypt := NewArgonCrypt()

		if crypt.mk != nil {
			t.Fatal("failed to initialize master key empty")
		}

		expectedKey := []byte("master-key")

		crypt.ConfigMK(expectedKey)

		if !reflect.DeepEqual(crypt.mk, expectedKey) {
			t.Fatalf("expected key: %v and got: %v", expectedKey, crypt.mk)
		}
	})
}

func newConfiguredCrypt(masterKey string) *ArgonCrypt {
	crypt := NewArgonCrypt()
	crypt.ConfigMK([]byte(masterKey))
	return crypt
}

func TestEncrypt(t *testing.T) {
	t.Run("should encrypt a secret successfully", func(t *testing.T) {
		crypt := newConfiguredCrypt("master-password")
		secret := domain.Secret{
			Name:      "test-secret",
			Password:  []byte("s3cur3p@ss"),
			CreatedAt: time.Now().Truncate(time.Second),
		}

		encrypted, err := crypt.Encrypt(secret)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(encrypted.Data) == 0 {
			t.Fatal("expected non-empty cipher data")
		}

		if encrypted.Salt == [MSK_SALT_SIZE]byte{} {
			t.Fatal("expected non-zero salt")
		}

		if encrypted.Nonce == [MSK_NONCE_SIZE]byte{} {
			t.Fatal("expected non-zero nonce")
		}
	})

	t.Run("should produce different cipher data for same input", func(t *testing.T) {
		crypt := newConfiguredCrypt("master-password")
		secret := domain.Secret{
			Name:      "test-secret",
			Password:  []byte("s3cur3p@ss"),
			CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		enc1, err := crypt.Encrypt(secret)
		if err != nil {
			t.Fatalf("first encrypt failed: %v", err)
		}

		enc2, err := crypt.Encrypt(secret)
		if err != nil {
			t.Fatalf("second encrypt failed: %v", err)
		}

		if reflect.DeepEqual(enc1.Data, enc2.Data) {
			t.Fatal("expected different cipher data due to random salt/nonce")
		}
	})

	t.Run("should return error when master key is not configured", func(t *testing.T) {
		crypt := NewArgonCrypt()
		secret := domain.Secret{
			Name:      "test",
			Password:  []byte("pass"),
			CreatedAt: time.Now(),
		}

		_, err := crypt.Encrypt(secret)
		if err == nil {
			t.Fatal("expected error when master key is empty")
		}
	})
}

func TestDecrypt(t *testing.T) {
	t.Run("should decrypt an encrypted secret correctly (round-trip)", func(t *testing.T) {
		crypt := newConfiguredCrypt("master-password")
		secret := domain.Secret{
			Name:      "my-secret",
			Password:  []byte("p@ssw0rd!"),
			CreatedAt: time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC),
		}

		encrypted, err := crypt.Encrypt(secret)
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		cipherData := buildCipherData(encrypted.Salt, encrypted.Nonce, encrypted.Data)

		decrypted, err := crypt.Decrypt(cipherData)
		if err != nil {
			t.Fatalf("decrypt failed: %v", err)
		}

		if decrypted.Name != secret.Name {
			t.Fatalf("expected name %q, got %q", secret.Name, decrypted.Name)
		}

		if !reflect.DeepEqual(decrypted.Password, secret.Password) {
			t.Fatalf("expected password %v, got %v", secret.Password, decrypted.Password)
		}

		if !decrypted.CreatedAt.Equal(secret.CreatedAt) {
			t.Fatalf("expected createdAt %v, got %v", secret.CreatedAt, decrypted.CreatedAt)
		}
	})

	t.Run("should return ErrCorruptedFile when data is too short", func(t *testing.T) {
		crypt := newConfiguredCrypt("master-password")
		shortData := []byte("MSK")

		_, err := crypt.Decrypt(shortData)
		if err == nil {
			t.Fatal("expected error for short data")
		}

		if !errors.Is(err, ErrCorruptedFile) {
			t.Fatalf("expected ErrCorruptedFile, got %v", err)
		}
	})

	t.Run("should return ErrCorruptedFile when magic value is wrong", func(t *testing.T) {
		crypt := newConfiguredCrypt("master-password")

		data := make([]byte, MSK_HEADER_SIZE+16)
		copy(data[:3], "BAD")
		data[3] = MSK_FILE_VERSION

		_, err := crypt.Decrypt(data)
		t.Log(err)
		if err == nil {
			t.Fatal("expected error for wrong magic value")
		}

		if !errors.Is(err, ErrCorruptedFile) {
			t.Fatalf("expected ErrCorruptedFile, got %v", err)
		}
	})

	t.Run("should return ErrUnsupportedFileVersion when version is wrong", func(t *testing.T) {
		crypt := newConfiguredCrypt("master-password")
		data := make([]byte, MSK_HEADER_SIZE+16)
		copy(data[:3], MSK_MAGIC_VALUE)
		data[3] = 99

		_, err := crypt.Decrypt(data)
		if err == nil {
			t.Fatal("expected error for unsupported version")
		}

		if !errors.Is(err, ErrUnsupportedFileVersion) {
			t.Fatalf("expected ErrUnsupportedFileVersion, got %v", err)
		}
	})

	t.Run("should return ErrDecryption when master key is wrong", func(t *testing.T) {
		crypt := newConfiguredCrypt("correct-password")
		secret := domain.Secret{
			Name:      "test",
			Password:  []byte("pass"),
			CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		encrypted, err := crypt.Encrypt(secret)
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		cipherData := buildCipherData(encrypted.Salt, encrypted.Nonce, encrypted.Data)

		wrongCrypt := newConfiguredCrypt("wrong-password")
		_, err = wrongCrypt.Decrypt(cipherData)
		if err == nil {
			t.Fatal("expected error with wrong master key")
		}

		if !errors.Is(err, ErrDecryption) {
			t.Fatalf("expected ErrDecryption, got %v", err)
		}
	})

	t.Run("should return ErrDecryption when cipher data is tampered", func(t *testing.T) {
		crypt := newConfiguredCrypt("master-password")
		secret := domain.Secret{
			Name:      "test",
			Password:  []byte("pass"),
			CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		encrypted, err := crypt.Encrypt(secret)
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		cipherData := buildCipherData(encrypted.Salt, encrypted.Nonce, encrypted.Data)

		// Flip a byte in the cipher data portion
		cipherData[MSK_HEADER_SIZE] ^= 0xFF

		_, err = crypt.Decrypt(cipherData)
		if err == nil {
			t.Fatal("expected error with tampered cipher data")
		}

		if !errors.Is(err, ErrDecryption) {
			t.Fatalf("expected ErrDecryption, got %v", err)
		}
	})

	t.Run("should return error when master key is empty", func(t *testing.T) {
		crypt := NewArgonCrypt()
		data := make([]byte, MSK_HEADER_SIZE+16)
		copy(data[:3], MSK_MAGIC_VALUE)
		data[3] = MSK_FILE_VERSION

		_, err := crypt.Decrypt(data)
		if err == nil {
			t.Fatal("expected error when master key is empty")
		}
	})
}
