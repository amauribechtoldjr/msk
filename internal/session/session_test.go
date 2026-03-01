package session

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestSession(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	sessionPathOverride = filepath.Join(tmpDir, "session")
	t.Cleanup(func() { sessionPathOverride = "" })
}

func newTestSession(t *testing.T) *Session {
	t.Helper()
	s, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	return s
}

func TestNew(t *testing.T) {
	t.Run("returns Session with resolved path", func(t *testing.T) {
		setupTestSession(t)
		s, err := New()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if s == nil {
			t.Fatal("expected non-nil Session")
		}
		if s.path == "" {
			t.Fatal("expected non-empty path")
		}
	})
}

func TestCreateAndLoad(t *testing.T) {
	t.Run("round-trip: Load returns same master key", func(t *testing.T) {
		setupTestSession(t)
		s := newTestSession(t)

		mk := []byte("this-is-a-32-byte-master-key-!!")
		token, err := s.Create(mk)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if token == "" {
			t.Fatal("expected non-empty token")
		}

		got, err := s.Load(token)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}
		if !bytes.Equal(got, mk) {
			t.Fatalf("Load returned %x, want %x", got, mk)
		}
	})

	t.Run("Create overwrites previous session file", func(t *testing.T) {
		setupTestSession(t)
		s := newTestSession(t)

		mk1 := []byte("first-master-key-bytes-here-32!!")
		token1, _ := s.Create(mk1)

		mk2 := []byte("second-master-key-bytes-here-32!")
		token2, _ := s.Create(mk2)

		got, err := s.Load(token2)
		if err != nil {
			t.Fatalf("Load with token2 failed: %v", err)
		}
		if !bytes.Equal(got, mk2) {
			t.Fatalf("expected mk2, got %x", got)
		}

		_, err = s.Load(token1)
		if err == nil {
			t.Fatal("expected Load with stale token1 to fail")
		}
	})

	t.Run("Load returns error for wrong token", func(t *testing.T) {
		setupTestSession(t)
		s := newTestSession(t)

		mk := []byte("master-key-for-wrong-token-test!")
		_, _ = s.Create(mk)

		wrongToken := "aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899"
		_, err := s.Load(wrongToken)
		if err == nil {
			t.Fatal("expected error with wrong token")
		}
	})
}

func TestCreateTokenFormat(t *testing.T) {
	t.Run("token is 64-char hex string", func(t *testing.T) {
		setupTestSession(t)
		s := newTestSession(t)
		mk := []byte("any-mk-bytes-32-bytes-long-here!")
		token, _ := s.Create(mk)
		if len(token) != 64 {
			t.Fatalf("expected 64-char token, got len=%d", len(token))
		}
	})

	t.Run("consecutive Create calls produce different tokens", func(t *testing.T) {
		setupTestSession(t)
		s := newTestSession(t)
		mk := []byte("any-mk-bytes-32-bytes-long-here!")
		t1, _ := s.Create(mk)
		t2, _ := s.Create(mk)
		if t1 == t2 {
			t.Fatal("expected different tokens")
		}
	})
}

func TestExpiry(t *testing.T) {
	t.Run("Load returns ErrSessionExpired for expired session", func(t *testing.T) {
		setupTestSession(t)
		s := newTestSession(t)
		mk := []byte("master-key-expiry-test-32-bytes!")
		token, _ := s.Create(mk)

		data, _ := os.ReadFile(s.path)
		pastExpiry := int64(1)
		for i := 0; i < 8; i++ {
			data[12+i] = byte(pastExpiry >> (56 - 8*i))
		}
		os.WriteFile(s.path, data, 0o600)

		_, err := s.Load(token)
		if err != ErrSessionExpired {
			t.Fatalf("expected ErrSessionExpired, got %v", err)
		}
	})
}

func TestRefresh(t *testing.T) {
	t.Run("extends the expiry timestamp", func(t *testing.T) {
		setupTestSession(t)
		s := newTestSession(t)
		mk := []byte("master-key-refresh-test-32-byt!!")
		_, _ = s.Create(mk)

		data, _ := os.ReadFile(s.path)
		var before int64
		for i := 0; i < 8; i++ {
			before = (before << 8) | int64(data[12+i])
		}

		time.Sleep(1 * time.Second)

		if err := s.Refresh(); err != nil {
			t.Fatalf("Refresh failed: %v", err)
		}

		data2, _ := os.ReadFile(s.path)
		var after int64
		for i := 0; i < 8; i++ {
			after = (after << 8) | int64(data2[12+i])
		}
		if after <= before {
			t.Fatalf("expected expiry to increase: before=%d after=%d", before, after)
		}
	})

	t.Run("preserves encrypted master key after refresh", func(t *testing.T) {
		setupTestSession(t)
		s := newTestSession(t)
		mk := []byte("master-key-refresh-preserve-32!!")
		token, _ := s.Create(mk)

		s.Refresh()

		got, err := s.Load(token)
		if err != nil {
			t.Fatalf("Load after Refresh failed: %v", err)
		}
		if !bytes.Equal(got, mk) {
			t.Fatalf("mk changed after Refresh")
		}
	})

	t.Run("returns error when no session file", func(t *testing.T) {
		setupTestSession(t)
		s := newTestSession(t)
		if err := s.Refresh(); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestDestroy(t *testing.T) {
	t.Run("removes the session file", func(t *testing.T) {
		setupTestSession(t)
		s := newTestSession(t)
		mk := []byte("master-key-destroy-test-32-byte!")
		_, _ = s.Create(mk)
		s.Destroy()
		if _, err := os.Stat(s.path); !os.IsNotExist(err) {
			t.Fatal("expected file gone after Destroy")
		}
	})

	t.Run("idempotent when file does not exist", func(t *testing.T) {
		setupTestSession(t)
		s := newTestSession(t)
		if err := s.Destroy(); err != nil {
			t.Fatalf("Destroy on nonexistent file: %v", err)
		}
	})
}

func TestIsActive(t *testing.T) {
	t.Run("true for fresh session", func(t *testing.T) {
		setupTestSession(t)
		s := newTestSession(t)
		mk := []byte("master-key-isactive-test-32-by!!")
		token, _ := s.Create(mk)
		if !s.IsActive(token) {
			t.Fatal("expected true")
		}
	})

	t.Run("false when no session file", func(t *testing.T) {
		setupTestSession(t)
		s := newTestSession(t)
		if s.IsActive("anytoken") {
			t.Fatal("expected false")
		}
	})

	t.Run("false when expired", func(t *testing.T) {
		setupTestSession(t)
		s := newTestSession(t)
		mk := []byte("master-key-isactive-expire-32!!!")
		token, _ := s.Create(mk)
		data, _ := os.ReadFile(s.path)
		pastExpiry := int64(1)
		for i := 0; i < 8; i++ {
			data[12+i] = byte(pastExpiry >> (56 - 8*i))
		}
		os.WriteFile(s.path, data, 0o600)
		if s.IsActive(token) {
			t.Fatal("expected false for expired")
		}
	})

	t.Run("false for wrong token", func(t *testing.T) {
		setupTestSession(t)
		s := newTestSession(t)
		mk := []byte("master-key-isactive-wrong-token!")
		_, _ = s.Create(mk)
		if s.IsActive("0000000000000000000000000000000000000000000000000000000000000000") {
			t.Fatal("expected false for wrong token")
		}
	})
}
