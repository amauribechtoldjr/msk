package format

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/amauribechtoldjr/msk/internal/domain"
)

func TestMarshalUnmarshalSecret(t *testing.T) {
	t.Run("should round-trip a secret correctly", func(t *testing.T) {
		secret := domain.Secret{
			Name:     "my-secret",
			Password: []byte("p@ssw0rd!"),
		}

		data := MarshalSecret(secret)
		got, err := UnmarshalSecret(data)
		if err != nil {
			t.Fatalf("failed to unmarshal secret: %v", err)
		}

		if got.Name != secret.Name {
			t.Fatalf("expected name %q, got %q", secret.Name, got.Name)
		}

		if !bytes.Equal(got.Password, secret.Password) {
			t.Fatalf("expected password %v, got %v", secret.Password, got.Password)
		}
	})

	t.Run("should handle empty name", func(t *testing.T) {
		secret := domain.Secret{
			Name:     "",
			Password: []byte("pass"),
		}

		data := MarshalSecret(secret)
		got, err := UnmarshalSecret(data)
		if err != nil {
			t.Fatalf("failed to unmarshal secret: %v", err)
		}

		if got.Name != "" {
			t.Fatalf("expected empty name, got %q", got.Name)
		}

		if !bytes.Equal(got.Password, secret.Password) {
			t.Fatalf("expected password %v, got %v", secret.Password, got.Password)
		}
	})

	t.Run("should handle empty password", func(t *testing.T) {
		secret := domain.Secret{
			Name:     "test",
			Password: []byte{},
		}

		data := MarshalSecret(secret)
		got, err := UnmarshalSecret(data)
		if err != nil {
			t.Fatalf("failed to unmarshal secret: %v", err)
		}

		if got.Name != secret.Name {
			t.Fatalf("expected name %q, got %q", secret.Name, got.Name)
		}

		if len(got.Password) != 0 {
			t.Fatalf("expected empty password, got %v", got.Password)
		}
	})

	t.Run("should handle binary password data", func(t *testing.T) {
		secret := domain.Secret{
			Name:     "binary-test",
			Password: []byte{0x00, 0xFF, 0x01, 0xFE},
		}

		data := MarshalSecret(secret)
		got, err := UnmarshalSecret(data)
		if err != nil {
			t.Fatalf("failed to unmarshal secret: %v", err)
		}

		if !bytes.Equal(got.Password, secret.Password) {
			t.Fatalf("expected password %v, got %v", secret.Password, got.Password)
		}
	})

	t.Run("should produce deterministic output", func(t *testing.T) {
		secret := domain.Secret{
			Name:     "test",
			Password: []byte("pass"),
		}

		data1 := MarshalSecret(secret)
		data2 := MarshalSecret(secret)

		if !reflect.DeepEqual(data1, data2) {
			t.Fatal("expected identical marshal output for same input")
		}
	})
}

func TestMarshalSecretFormat(t *testing.T) {
	t.Run("should produce correct binary layout", func(t *testing.T) {
		secret := domain.Secret{
			Name:     "ab",
			Password: []byte("xyz"),
		}

		data := MarshalSecret(secret)

		expectedLen := MSK_NAME_LENGTH_SIZE + 2 + MSK_PASSWORD_LENGTH_SIZE + 3
		if len(data) != expectedLen {
			t.Fatalf("expected length %d, got %d", expectedLen, len(data))
		}

		if data[0] != 0x00 || data[1] != 0x02 {
			t.Fatalf("expected name length bytes [0x00, 0x02], got [0x%02x, 0x%02x]", data[0], data[1])
		}

		if string(data[2:4]) != "ab" {
			t.Fatalf("expected name %q, got %q", "ab", string(data[2:4]))
		}

		if data[4] != 0x00 || data[5] != 0x03 {
			t.Fatalf("expected pass length bytes [0x00, 0x03], got [0x%02x, 0x%02x]", data[4], data[5])
		}

		if string(data[6:9]) != "xyz" {
			t.Fatalf("expected password %q, got %q", "xyz", string(data[6:9]))
		}
	})
}

func TestMarshalFile(t *testing.T) {
	makeSalt := func() []byte {
		s := make([]byte, MSK_SALT_SIZE)
		for i := range s {
			s[i] = byte(i + 1)
		}
		return s
	}

	makeNonce := func() []byte {
		n := make([]byte, MSK_NONCE_SIZE)
		for i := range n {
			n[i] = byte(i + 0xA0)
		}
		return n
	}

	t.Run("should produce correct binary layout", func(t *testing.T) {
		salt := makeSalt()
		nonce := makeNonce()
		data := []byte("ciphertext")

		file, err := MarshalFile(salt, nonce, data)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectedLen := MSK_HEADER_SIZE + len(data)
		if len(file) != expectedLen {
			t.Fatalf("expected length %d, got %d", expectedLen, len(file))
		}

		if string(file[:MSK_MAGIC_SIZE]) != MSK_MAGIC_VALUE {
			t.Fatalf("expected magic %q, got %q", MSK_MAGIC_VALUE, string(file[:MSK_MAGIC_SIZE]))
		}

		if file[MSK_MAGIC_SIZE] != MSK_FILE_VERSION {
			t.Fatalf("expected version %d, got %d", MSK_FILE_VERSION, file[MSK_MAGIC_SIZE])
		}

		offset := MSK_MAGIC_SIZE + MSK_VERSION_SIZE

		if !bytes.Equal(file[offset:offset+MSK_SALT_SIZE], salt) {
			t.Fatal("salt mismatch")
		}
		offset += MSK_SALT_SIZE

		if !bytes.Equal(file[offset:offset+MSK_NONCE_SIZE], nonce) {
			t.Fatal("nonce mismatch")
		}
		offset += MSK_NONCE_SIZE

		if !bytes.Equal(file[offset:], data) {
			t.Fatal("data mismatch")
		}
	})

	t.Run("should handle nil data", func(t *testing.T) {
		file, err := MarshalFile(makeSalt(), makeNonce(), nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(file) != MSK_HEADER_SIZE {
			t.Fatalf("expected length %d, got %d", MSK_HEADER_SIZE, len(file))
		}
	})

	t.Run("should return error for invalid salt size", func(t *testing.T) {
		_, err := MarshalFile([]byte("short"), makeNonce(), []byte("data"))
		if err == nil {
			t.Fatal("expected error for invalid salt size")
		}
	})

	t.Run("should return error for invalid nonce size", func(t *testing.T) {
		_, err := MarshalFile(makeSalt(), []byte("short"), []byte("data"))
		if err == nil {
			t.Fatal("expected error for invalid nonce size")
		}
	})
}

func TestUnmarshalFile(t *testing.T) {
	makeSalt := func() []byte {
		s := make([]byte, MSK_SALT_SIZE)
		for i := range s {
			s[i] = byte(i + 1)
		}
		return s
	}

	makeNonce := func() []byte {
		n := make([]byte, MSK_NONCE_SIZE)
		for i := range n {
			n[i] = byte(i + 0xA0)
		}
		return n
	}

	t.Run("should round-trip with MarshalFile", func(t *testing.T) {
		salt := makeSalt()
		nonce := makeNonce()
		data := []byte("encrypted-payload")

		file, err := MarshalFile(salt, nonce, data)
		if err != nil {
			t.Fatalf("marshal failed: %v", err)
		}

		gotSalt, gotNonce, gotData, err := UnmarshalFile(file)
		if err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}

		if !bytes.Equal(gotSalt, salt) {
			t.Fatal("salt mismatch")
		}

		if !bytes.Equal(gotNonce, nonce) {
			t.Fatal("nonce mismatch")
		}

		if !bytes.Equal(gotData, data) {
			t.Fatal("data mismatch")
		}
	})

	t.Run("should return ErrCorruptedFile when data is too short", func(t *testing.T) {
		_, _, _, err := UnmarshalFile([]byte("MSK"))
		if err != ErrCorruptedFile {
			t.Fatalf("expected ErrCorruptedFile, got %v", err)
		}
	})

	t.Run("should return ErrCorruptedFile when magic is wrong", func(t *testing.T) {
		data := make([]byte, MSK_HEADER_SIZE+10)
		copy(data[:3], "BAD")
		data[3] = MSK_FILE_VERSION

		_, _, _, err := UnmarshalFile(data)
		if err != ErrCorruptedFile {
			t.Fatalf("expected ErrCorruptedFile, got %v", err)
		}
	})

	t.Run("should return ErrUnsupportedFileVersion when version is wrong", func(t *testing.T) {
		data := make([]byte, MSK_HEADER_SIZE+10)
		copy(data[:3], MSK_MAGIC_VALUE)
		data[3] = 99

		_, _, _, err := UnmarshalFile(data)
		if err != ErrUnsupportedFileVersion {
			t.Fatalf("expected ErrUnsupportedFileVersion, got %v", err)
		}
	})

	t.Run("should return empty data when file has header only", func(t *testing.T) {
		file, err := MarshalFile(makeSalt(), makeNonce(), nil)
		if err != nil {
			t.Fatalf("marshal failed: %v", err)
		}

		_, _, gotData, err := UnmarshalFile(file)
		if err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}

		if len(gotData) != 0 {
			t.Fatalf("expected empty data, got %v", gotData)
		}
	})

	t.Run("should return ErrCorruptedFile for empty input", func(t *testing.T) {
		_, _, _, err := UnmarshalFile([]byte{})
		if err != ErrCorruptedFile {
			t.Fatalf("expected ErrCorruptedFile, got %v", err)
		}
	})

	t.Run("should return ErrCorruptedFile for nil input", func(t *testing.T) {
		_, _, _, err := UnmarshalFile(nil)
		if err != ErrCorruptedFile {
			t.Fatalf("expected ErrCorruptedFile, got %v", err)
		}
	})
}
