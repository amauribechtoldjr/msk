package vault

import (
	"errors"
	"reflect"
	"testing"

	"github.com/amauribechtoldjr/msk/internal/domain"
	"github.com/amauribechtoldjr/msk/internal/format"
)

func TestNewMSKVault(t *testing.T) {
	t.Run("should initialize the struct correctly", func(t *testing.T) {
		var crypt Vault = NewMSKVault()
		_, ok := crypt.(*MSKVault)

		if !ok {
			t.Fatal("expected and variable of type MSKVault")
		}
	})
}

func TestConfigMk(t *testing.T) {
	t.Run("should set the master key correctly", func(t *testing.T) {
		crypt := NewMSKVault()

		if crypt.mk != nil {
			t.Fatal("failed to initialize master key empty")
		}

		crypt.ConfigMK([]byte("master-key"))
		if crypt.mk == nil {
			t.Fatal("expected mk to be set after ConfigMK")
		}

		buffer, err := crypt.mk.Open()
		if err != nil {
			t.Fatal("failed to open the master key enclave buffer")
		}

		expectedKey := []byte("master-key")

		if !reflect.DeepEqual(buffer.Bytes(), expectedKey) {
			t.Fatalf("expected key: %v and got: %v", expectedKey, crypt.mk)
		}
	})
}

func newConfiguredCrypt(masterKey string) *MSKVault {
	crypt := NewMSKVault()
	crypt.ConfigMK([]byte(masterKey))
	return crypt
}

func TestEncrypt(t *testing.T) {
	t.Run("should encrypt a secret successfully", func(t *testing.T) {
		crypt := newConfiguredCrypt("master-password")
		secret := domain.Secret{
			Name:     "test-secret",
			Password: []byte("s3cur3p@ss"),
		}

		encrypted, err := crypt.EncryptSecret(secret)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		salt, nonce, data, err := format.UnmarshalFile(encrypted)
		if err != nil {
			t.Fatalf("failed to unmarshal encrypted output: %v", err)
		}

		if len(data) == 0 {
			t.Fatal("expected non-empty cipher data")
		}

		if len(salt) != format.MSK_SALT_SIZE {
			t.Fatal("expected valid salt")
		}

		if len(nonce) != format.MSK_NONCE_SIZE {
			t.Fatal("expected valid nonce")
		}
	})

	t.Run("should produce different cipher data for same input", func(t *testing.T) {
		crypt := newConfiguredCrypt("master-password")
		secret := domain.Secret{
			Name:     "test-secret",
			Password: []byte("s3cur3p@ss"),
		}

		enc1, err := crypt.EncryptSecret(secret)
		if err != nil {
			t.Fatalf("first encrypt failed: %v", err)
		}

		enc2, err := crypt.EncryptSecret(secret)
		if err != nil {
			t.Fatalf("second encrypt failed: %v", err)
		}

		if reflect.DeepEqual(enc1, enc2) {
			t.Fatal("expected different cipher data due to random salt/nonce")
		}
	})

	t.Run("should return error when master key is not configured", func(t *testing.T) {
		crypt := NewMSKVault()
		secret := domain.Secret{
			Name:     "test",
			Password: []byte("pass"),
		}

		_, err := crypt.EncryptSecret(secret)
		if err == nil {
			t.Fatal("expected error when master key is empty")
		}
	})
}

func TestDecrypt(t *testing.T) {
	t.Run("should decrypt an encrypted secret correctly (round-trip)", func(t *testing.T) {
		crypt := newConfiguredCrypt("master-password")
		secret := domain.Secret{
			Name:     "my-secret",
			Password: []byte("p@ssw0rd!"),
		}

		encrypted, err := crypt.EncryptSecret(secret)
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		decrypted, err := crypt.DecryptSecret(encrypted)
		if err != nil {
			t.Fatalf("decrypt failed: %v", err)
		}

		if decrypted.Name != secret.Name {
			t.Fatalf("expected name %q, got %q", secret.Name, decrypted.Name)
		}

		if !reflect.DeepEqual(decrypted.Password, secret.Password) {
			t.Fatalf("expected password %v, got %v", secret.Password, decrypted.Password)
		}
	})

	t.Run("should return ErrCorruptedFile when data is too short", func(t *testing.T) {
		crypt := newConfiguredCrypt("master-password")
		shortData := []byte("MSK")

		_, err := crypt.DecryptSecret(shortData)
		if err == nil {
			t.Fatal("expected error for short data")
		}

		if !errors.Is(err, format.ErrCorruptedFile) {
			t.Fatalf("expected ErrCorruptedFile, got %v", err)
		}
	})

	t.Run("should return ErrCorruptedFile when magic value is wrong", func(t *testing.T) {
		crypt := newConfiguredCrypt("master-password")

		data := make([]byte, format.MSK_HEADER_SIZE+16)
		copy(data[:3], "BAD")
		data[3] = format.MSK_FILE_VERSION

		_, err := crypt.DecryptSecret(data)
		if err == nil {
			t.Fatal("expected error for wrong magic value")
		}

		if !errors.Is(err, format.ErrCorruptedFile) {
			t.Fatalf("expected ErrCorruptedFile, got %v", err)
		}
	})

	t.Run("should return ErrUnsupportedFileVersion when version is wrong", func(t *testing.T) {
		crypt := newConfiguredCrypt("master-password")
		data := make([]byte, format.MSK_HEADER_SIZE+16)
		copy(data[:3], format.MSK_MAGIC_VALUE)
		data[3] = 99

		_, err := crypt.DecryptSecret(data)
		if err == nil {
			t.Fatal("expected error for unsupported version")
		}

		if !errors.Is(err, format.ErrUnsupportedFileVersion) {
			t.Fatalf("expected ErrUnsupportedFileVersion, got %v", err)
		}
	})

	t.Run("should return ErrDecryption when master key is wrong", func(t *testing.T) {
		crypt := newConfiguredCrypt("correct-password")
		secret := domain.Secret{
			Name:     "test",
			Password: []byte("pass"),
		}

		encrypted, err := crypt.EncryptSecret(secret)
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		wrongCrypt := newConfiguredCrypt("wrong-password")
		_, err = wrongCrypt.DecryptSecret(encrypted)
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
			Name:     "test",
			Password: []byte("pass"),
		}

		encrypted, err := crypt.EncryptSecret(secret)
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		// Flip a byte in the cipher data portion
		encrypted[format.MSK_HEADER_SIZE] ^= 0xFF

		_, err = crypt.DecryptSecret(encrypted)
		if err == nil {
			t.Fatal("expected error with tampered cipher data")
		}

		if !errors.Is(err, ErrDecryption) {
			t.Fatalf("expected ErrDecryption, got %v", err)
		}
	})

	t.Run("should return error when master key is empty", func(t *testing.T) {
		crypt := NewMSKVault()
		data := make([]byte, format.MSK_HEADER_SIZE+16)
		copy(data[:3], format.MSK_MAGIC_VALUE)
		data[3] = format.MSK_FILE_VERSION

		_, err := crypt.DecryptSecret(data)
		if err == nil {
			t.Fatal("expected error when master key is empty")
		}
	})
}
