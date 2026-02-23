package app

import (
	"errors"
	"reflect"
	"testing"

	"github.com/amauribechtoldjr/msk/internal/encryption"
	"github.com/amauribechtoldjr/msk/internal/storage"
)

func newTestService(t *testing.T, masterKey string) *MSKService {
	t.Helper()

	store, err := storage.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	crypto := encryption.NewArgonCrypt()
	crypto.ConfigMK([]byte(masterKey))

	return NewMSKService(store, crypto)
}

func TestNewMSKService(t *testing.T) {
	t.Run("should return a properly initialized service", func(t *testing.T) {
		service := newTestService(t, "master-key")

		if service == nil {
			t.Fatal("expected non-nil service")
		}
	})
}

func TestConfigMK(t *testing.T) {
	t.Run("should allow decryption after reconfiguring master key", func(t *testing.T) {
		service := newTestService(t, "first-key")

		err := service.AddSecret("secret", []byte("password"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		service.ConfigMK([]byte("wrong-key"))
		_, err = service.GetSecret("secret")
		if err == nil {
			t.Fatal("expected error with wrong master key")
		}

		service.ConfigMK([]byte("first-key"))
		password, err := service.GetSecret("secret")
		if err != nil {
			t.Fatalf("expected no error after restoring key, got %v", err)
		}

		if !reflect.DeepEqual(password, []byte("password")) {
			t.Fatalf("expected password %q, got %q", "password", password)
		}
	})
}

func TestAddSecret(t *testing.T) {
	t.Run("should add secret successfully", func(t *testing.T) {
		service := newTestService(t, "master-key")

		err := service.AddSecret("my-secret", []byte("p@ssword"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("should return ErrSecretExists when secret already exists", func(t *testing.T) {
		service := newTestService(t, "master-key")

		err := service.AddSecret("duplicate", []byte("pass"))
		if err != nil {
			t.Fatalf("first add failed: %v", err)
		}

		err = service.AddSecret("duplicate", []byte("pass2"))
		if !errors.Is(err, ErrSecretExists) {
			t.Fatalf("expected ErrSecretExists, got %v", err)
		}
	})

	t.Run("should return error when encryption fails with empty master key", func(t *testing.T) {
		store, err := storage.NewStore(t.TempDir())
		if err != nil {
			t.Fatalf("failed to create store: %v", err)
		}

		crypto := encryption.NewArgonCrypt()
		service := NewMSKService(store, crypto)

		err = service.AddSecret("secret", []byte("pass"))
		if err == nil {
			t.Fatal("expected error when master key is not configured")
		}
	})
}

func TestGetSecret(t *testing.T) {
	t.Run("should return decrypted password successfully", func(t *testing.T) {
		service := newTestService(t, "master-key")
		expected := []byte("s3cur3p@ss")
		inputPass := []byte("s3cur3p@ss")

		err := service.AddSecret("my-secret", inputPass)
		if err != nil {
			t.Fatalf("add failed: %v", err)
		}

		password, err := service.GetSecret("my-secret")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !reflect.DeepEqual(password, expected) {
			t.Fatalf("expected password %q, got %q", expected, password)
		}
	})

	t.Run("should return ErrSecretNotFound when secret does not exist", func(t *testing.T) {
		service := newTestService(t, "master-key")

		_, err := service.GetSecret("missing")
		if !errors.Is(err, ErrSecretNotFound) {
			t.Fatalf("expected ErrSecretNotFound, got %v", err)
		}
	})

	t.Run("should return error when decryption fails with wrong key", func(t *testing.T) {
		service := newTestService(t, "correct-key")

		err := service.AddSecret("secret", []byte("pass"))
		if err != nil {
			t.Fatalf("add failed: %v", err)
		}

		service.ConfigMK([]byte("wrong-key"))

		_, err = service.GetSecret("secret")
		if err == nil {
			t.Fatal("expected error with wrong master key")
		}
	})
}

func TestDeleteSecret(t *testing.T) {
	t.Run("should delete secret successfully", func(t *testing.T) {
		service := newTestService(t, "master-key")

		err := service.AddSecret("to-delete", []byte("pass"))
		if err != nil {
			t.Fatalf("add failed: %v", err)
		}

		err = service.DeleteSecret("to-delete")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		_, err = service.GetSecret("to-delete")
		if !errors.Is(err, ErrSecretNotFound) {
			t.Fatalf("expected ErrSecretNotFound after delete, got %v", err)
		}
	})

	t.Run("should return error when secret does not exist", func(t *testing.T) {
		service := newTestService(t, "master-key")

		err := service.DeleteSecret("nonexistent")
		if err == nil {
			t.Fatal("expected error when deleting nonexistent secret")
		}
	})

}

func TestUpdateSecret(t *testing.T) {
	t.Run("should update secret successfully", func(t *testing.T) {
		service := newTestService(t, "master-key")

		err := service.AddSecret("to-update", []byte("old-pass"))
		if err != nil {
			t.Fatalf("add failed: %v", err)
		}

		err = service.UpdateSecret("to-update", []byte("new-pass"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		password, err := service.GetSecret("to-update")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !reflect.DeepEqual(password, []byte("new-pass")) {
			t.Fatalf("expected password %q, got %q", "new-pass", password)
		}
	})

	t.Run("should return ErrSecretNotFound when secret does not exist", func(t *testing.T) {
		service := newTestService(t, "master-key")

		err := service.UpdateSecret("nonexistent", []byte("pass"))
		if !errors.Is(err, ErrSecretNotFound) {
			t.Fatalf("expected ErrSecretNotFound, got %v", err)
		}
	})

	t.Run("should not return old password after update", func(t *testing.T) {
		service := newTestService(t, "master-key")

		err := service.AddSecret("to-update", []byte("old-pass"))
		if err != nil {
			t.Fatalf("add failed: %v", err)
		}

		err = service.UpdateSecret("to-update", []byte("new-pass"))
		if err != nil {
			t.Fatalf("update failed: %v", err)
		}

		password, err := service.GetSecret("to-update")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if reflect.DeepEqual(password, []byte("old-pass")) {
			t.Fatal("password should have changed after update")
		}
	})
}

func TestListSecrets(t *testing.T) {
	t.Run("should return list of secrets", func(t *testing.T) {
		service := newTestService(t, "master-key")

		err := service.AddSecret("secret-1", []byte("pass1"))
		if err != nil {
			t.Fatalf("add failed: %v", err)
		}

		err = service.AddSecret("secret-2", []byte("pass2"))
		if err != nil {
			t.Fatalf("add failed: %v", err)
		}

		files, err := service.ListSecrets()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(files) != 2 {
			t.Fatalf("expected 2 files, got %d", len(files))
		}
	})

	t.Run("should return empty slice when no secrets exist", func(t *testing.T) {
		service := newTestService(t, "master-key")

		files, err := service.ListSecrets()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(files) != 0 {
			t.Fatalf("expected 0 files, got %d", len(files))
		}
	})
}
