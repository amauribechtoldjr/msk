package storage

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/amauribechtoldjr/msk/internal/domain"
)

func initializeStore(t *testing.T) Store {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	return *store
}

func TestNewStore(t *testing.T) {
	t.Run("should create store with valid directory", func(t *testing.T) {
		store, err := NewStore(t.TempDir())
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		fileInfo, err := os.Stat(store.dir)
		if os.IsNotExist(err) {
			t.Fatalf("directory don't exists %v", err)
		} else if err != nil {
			t.Fatalf("error checking directory status %v", err)
		} else {
			if !fileInfo.IsDir() {
				t.Fatal("path exists but is not a directory")
			}
		}
	})

	t.Run("should not create store with path collision with an existing file", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "existing-file")
		if err != nil {
			panic(err)
		}
		tempFile.Close()
		defer os.Remove(tempFile.Name())

		storePath := filepath.Join(tempFile.Name(), "subdir")

		_, err = NewStore(storePath)
		if err != nil {
			return
		}

		t.Fatalf("store path created incorrectly: %v", err)
	})
}

func TestFileExists(t *testing.T) {
	t.Run("should return true if file exists", func(t *testing.T) {
		store := initializeStore(t)

		fileName := "existing-file"
		filePath := filepath.Join(store.dir, strings.Join([]string{fileName, ".msk"}, ""))
		err := os.WriteFile(filePath, []byte{}, 0o600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		exists := store.FileExists(fileName)

		if !exists {
			t.Fatal("should return true when file exists")
		}
	})

	t.Run("should return false if file does not exists", func(t *testing.T) {
		store := initializeStore(t)

		fileName := "existing-file"
		exists := store.FileExists(fileName)

		if exists {
			t.Fatal("should return true when file exists")
		}
	})
}

func TestGetFile(t *testing.T) {
	t.Run("should return file contents for an existing secret", func(t *testing.T) {
		store := initializeStore(t)

		expected := []byte("MSK\x01some-encrypted-data")
		err := os.WriteFile(filepath.Join(store.dir, "mysecret.msk"), expected, 0o600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		data, err := store.GetFile("mysecret")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !bytes.Equal(data, expected) {
			t.Fatalf("expected %q, got %q", expected, data)
		}
	})

	t.Run("should return ErrNotFound for a secret that does not exists", func(t *testing.T) {
		store := initializeStore(t)

		_, err := store.GetFile("doesnotexist")
		if err == nil {
			t.Fatal("expected an error, got nil")
		}

		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("should resolve names case-insensitively", func(t *testing.T) {
		store := initializeStore(t)

		expected := []byte("case-test-data")
		err := os.WriteFile(filepath.Join(store.dir, "mykey.msk"), expected, 0o600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		data, err := store.GetFile("MyKey")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !bytes.Equal(data, expected) {
			t.Fatalf("expected %q, got %q", expected, data)
		}
	})
}

func TestDeleteFile(t *testing.T) {
	t.Run("should delete existing file correctly", func(t *testing.T) {
		store := initializeStore(t)

		fileName := "existing-file"
		filePath := filepath.Join(store.dir, strings.Join([]string{fileName, ".msk"}, ""))
		err := os.WriteFile(filePath, []byte{}, 0o600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		err = store.DeleteFile(fileName)
		if err != nil {
			t.Fatalf("failed to delete file: %v", err)
		}

		_, err = os.Stat(filePath)
		if err == nil {
			t.Fatal("failed to delete file, it was retrieved successfully")
		}
	})

	t.Run("should return ErrNotFound when file does not exists", func(t *testing.T) {
		fileName := "does-not-exists"
		store := initializeStore(t)

		err := store.DeleteFile(fileName)
		if err != nil && !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected %v, got %v", ErrNotFound, err)
		}
	})

	t.Run("should resolve names case-insensitively", func(t *testing.T) {
		store := initializeStore(t)

		err := os.WriteFile(filepath.Join(store.dir, "mykey.msk"), []byte{}, 0o600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		err = store.DeleteFile("MyKey")
		if err != nil {
			t.Fatal("failed to delete file case-insensitively")
		}
	})
}

func TestGetFiles(t *testing.T) {
	t.Run("should return all .msk files", func(t *testing.T) {
		store := initializeStore(t)

		expectedFiles := []string{
			"file-1",
			"file-2",
		}

		for _, fileName := range expectedFiles {
			err := os.WriteFile(filepath.Join(store.dir, fileName+".msk"), []byte{}, 0o600)
			if err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}
		}

		files, err := store.GetFiles()
		if err != nil {
			t.Fatal("failed to retrieve existing files")
		}

		if len(expectedFiles) != len(files) {
			t.Fatal("failed to retrieve all existing files")
		}

		expectedMap := make(map[string]bool, len(expectedFiles))

		for _, fileName := range expectedFiles {
			expectedMap[fileName+".msk"] = true
		}

		for _, fileName := range files {
			if !expectedMap[fileName] {
				t.Fatalf("files array did not match, expected %s, got %s", expectedFiles, files)
			}
		}
	})

	t.Run("should only return .msk files", func(t *testing.T) {
		store := initializeStore(t)

		expectedFiles := []string{
			"file-1",
			"file-2",
		}

		for _, fileName := range expectedFiles {
			err := os.WriteFile(filepath.Join(store.dir, fileName+".msk"), []byte{}, 0o600)
			if err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}
		}

		err := os.WriteFile(filepath.Join(store.dir, "another.fake"), []byte{}, 0o600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		files, err := store.GetFiles()
		if err != nil {
			t.Fatal("failed to retrieve existing files")
		}

		if len(expectedFiles) != len(files) {
			t.Fatal("failed to retrieve all existing files")
		}

		expectedMap := make(map[string]bool, len(expectedFiles))

		for _, fileName := range expectedFiles {
			expectedMap[fileName+".msk"] = true
		}

		for _, fileName := range files {
			if !expectedMap[fileName] {
				t.Fatalf("files array did not match, expected %s, got %s", expectedFiles, files)
			}
		}
	})

	t.Run("should return empty string array when no files", func(t *testing.T) {
		store := initializeStore(t)

		err := os.WriteFile(filepath.Join(store.dir, "im-not-msk-file.fake"), []byte{}, 0o600)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		files, err := store.GetFiles()
		if err != nil {
			t.Fatal("failed to retrieve existing files")
		}

		if len(files) != 0 {
			t.Fatalf("expected files quantity: %v, got %v", 0, len(files))
		}
	})
}

func TestSaveFile(t *testing.T) {
	makeSalt := func() [16]byte {
		var s [16]byte
		for i := range s {
			s[i] = byte(i + 1)
		}
		return s
	}

	makeNonce := func() [12]byte {
		var n [12]byte
		for i := range n {
			n[i] = byte(i + 0xA0)
		}
		return n
	}

	t.Run("should create file with correct binary content", func(t *testing.T) {
		store := initializeStore(t)
		salt := makeSalt()
		nonce := makeNonce()
		data := []byte("encrypted-payload")

		secret := domain.EncryptedSecret{
			Salt:  salt,
			Nonce: nonce,
			Data:  data,
		}

		err := store.SaveFile(secret, "testsecret")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		raw, err := os.ReadFile(filepath.Join(store.dir, "testsecret.msk"))
		if err != nil {
			t.Fatalf("failed to read saved file: %v", err)
		}

		var expected []byte
		expected = append(expected, []byte("MSK")...)
		expected = append(expected, 0x01)
		expected = append(expected, salt[:]...)
		expected = append(expected, nonce[:]...)
		expected = append(expected, data...)

		if !bytes.Equal(raw, expected) {
			t.Fatalf("file content mismatch\nexpected: %x\ngot:      %x", expected, raw)
		}
	})

	t.Run("should store file with lowercased name", func(t *testing.T) {
		store := initializeStore(t)

		secret := domain.EncryptedSecret{
			Salt:  makeSalt(),
			Nonce: makeNonce(),
			Data:  []byte("data"),
		}

		err := store.SaveFile(secret, "MySecret")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expectedPath := filepath.Join(store.dir, "mysecret.msk")
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Fatal("expected file mysecret.msk but it does not exist")
		}
	})

	t.Run("should overwrite existing file", func(t *testing.T) {
		store := initializeStore(t)
		salt := makeSalt()
		nonce := makeNonce()

		first := domain.EncryptedSecret{Salt: salt, Nonce: nonce, Data: []byte("first")}
		second := domain.EncryptedSecret{Salt: salt, Nonce: nonce, Data: []byte("second")}

		if err := store.SaveFile(first, "overwrite"); err != nil {
			t.Fatalf("first save failed: %v", err)
		}
		if err := store.SaveFile(second, "overwrite"); err != nil {
			t.Fatalf("second save failed: %v", err)
		}

		raw, err := os.ReadFile(filepath.Join(store.dir, "overwrite.msk"))
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		if bytes.Contains(raw, []byte("first")) {
			t.Fatal("file still contains data from first write")
		}

		if !bytes.Contains(raw, []byte("second")) {
			t.Fatal("file does not contain data from second write")
		}
	})

	t.Run("should clean up temp file after success", func(t *testing.T) {
		store := initializeStore(t)

		secret := domain.EncryptedSecret{
			Salt:  makeSalt(),
			Nonce: makeNonce(),
			Data:  []byte("data"),
		}

		if err := store.SaveFile(secret, "cleanup"); err != nil {
			t.Fatalf("save failed: %v", err)
		}

		matches, err := filepath.Glob(filepath.Join(store.dir, "*.tmp"))
		if err != nil {
			t.Fatalf("glob failed: %v", err)
		}

		if len(matches) != 0 {
			t.Fatalf("expected no .tmp files, found: %v", matches)
		}
	})

	t.Run("should roundtrip with GetFile", func(t *testing.T) {
		store := initializeStore(t)
		salt := makeSalt()
		nonce := makeNonce()
		data := []byte("roundtrip-data")

		secret := domain.EncryptedSecret{
			Salt:  salt,
			Nonce: nonce,
			Data:  data,
		}

		if err := store.SaveFile(secret, "roundtrip"); err != nil {
			t.Fatalf("save failed: %v", err)
		}

		got, err := store.GetFile("roundtrip")
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}

		var expected []byte
		expected = append(expected, []byte("MSK")...)
		expected = append(expected, 0x01)
		expected = append(expected, salt[:]...)
		expected = append(expected, nonce[:]...)
		expected = append(expected, data...)

		if !bytes.Equal(got, expected) {
			t.Fatalf("roundtrip mismatch\nexpected: %x\ngot:      %x", expected, got)
		}
	})

	t.Run("should return error for unwritable directory", func(t *testing.T) {
		store := Store{dir: filepath.Join(t.TempDir(), "no", "such", "deep", "path")}

		secret := domain.EncryptedSecret{
			Salt:  makeSalt(),
			Nonce: makeNonce(),
			Data:  []byte("data"),
		}

		err := store.SaveFile(secret, "fail")
		if err == nil {
			t.Fatal("expected an error for unwritable directory, got nil")
		}
	})

	t.Run("should handle empty data field", func(t *testing.T) {
		store := initializeStore(t)
		salt := makeSalt()
		nonce := makeNonce()

		secret := domain.EncryptedSecret{
			Salt:  salt,
			Nonce: nonce,
			Data:  nil,
		}

		if err := store.SaveFile(secret, "emptydata"); err != nil {
			t.Fatalf("save failed: %v", err)
		}

		raw, err := os.ReadFile(filepath.Join(store.dir, "emptydata.msk"))
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		expectedLen := 3 + 1 + 16 + 12
		if len(raw) != expectedLen {
			t.Fatalf("expected file length %d, got %d", expectedLen, len(raw))
		}

		var expected []byte
		expected = append(expected, []byte("MSK")...)
		expected = append(expected, 0x01)
		expected = append(expected, salt[:]...)
		expected = append(expected, nonce[:]...)

		if !bytes.Equal(raw, expected) {
			t.Fatalf("file content mismatch\nexpected: %x\ngot:      %x", expected, raw)
		}
	})
}
