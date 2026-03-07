package generator

import (
	"strings"
	"testing"
)

func TestGeneratePassword_Length(t *testing.T) {
	lengths := []int{8, 16, 32, 64}
	for _, l := range lengths {
		pw, err := GeneratePassword(l, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pw) != l {
			t.Errorf("expected length %d, got %d", l, len(pw))
		}
	}
}

func TestGeneratePassword_DefaultLength(t *testing.T) {
	pw, err := GeneratePassword(0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pw) != 16 {
		t.Errorf("expected default length 16, got %d", len(pw))
	}
}

func TestGeneratePassword_NoSymbols(t *testing.T) {
	pw, err := GeneratePassword(100, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, b := range pw {
		if strings.ContainsRune(symbols, rune(b)) {
			t.Errorf("found symbol %q in no-symbols password", string(b))
		}
	}
}

func TestGeneratePassword_ValidCharacters(t *testing.T) {
	fullCharset := alphanumeric + symbols
	pw, err := GeneratePassword(200, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, b := range pw {
		if !strings.ContainsRune(fullCharset, rune(b)) {
			t.Errorf("invalid character %q in password", string(b))
		}
	}
}

func TestGeneratePassword_Uniqueness(t *testing.T) {
	pw1, err := GeneratePassword(32, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pw2, err := GeneratePassword(32, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(pw1) == string(pw2) {
		t.Error("two generated passwords should not be identical")
	}
}
