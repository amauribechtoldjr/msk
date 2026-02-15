package validator

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateName(t *testing.T) {
	t.Run("should return error for empty name", func(t *testing.T) {
		err := ValidateName("")
		if !errors.Is(err, ErrEmptyName) {
			t.Fatalf("expected ErrEmptyName, got %v", err)
		}
	})

	t.Run("should return error for too long name", func(t *testing.T) {
		err := ValidateName(strings.Repeat("T", 256))
		if !errors.Is(err, ErrNameTooLong) {
			t.Fatalf("expected ErrNameTooLong, got %v", err)
		}
	})

	t.Run("should return error for name with control characters", func(t *testing.T) {
		inputs := []string{
			"test\x00name",
			"test\x1Fname",
			"test\x7Fname",
			"my\tpassword",
			"my\npassword",
			"my\rpassword",
		}

		for _, input := range inputs {
			err := ValidateName(input)
			if !errors.Is(err, ErrControlCharacter) {
				t.Fatalf("ValidateName(%q): expected ErrControlCharacter, got %v", input, err)
			}
		}
	})

	t.Run("should return error for name with space", func(t *testing.T) {
		err := ValidateName("my password")
		if !errors.Is(err, ErrWhitespace) {
			t.Fatalf("expected ErrWhitespace, got %v", err)
		}
	})

	t.Run("should return error for name with path separators", func(t *testing.T) {
		inputs := []string{
			"path/to/secret",
			"path\\to\\secret",
			"trailing/",
			"/leading",
		}

		for _, input := range inputs {
			err := ValidateName(input)
			if !errors.Is(err, ErrPathSeparator) {
				t.Fatalf("ValidateName(%q): expected ErrPathSeparator, got %v", input, err)
			}
		}
	})

	t.Run("should return error for name with invalid characters", func(t *testing.T) {
		inputs := []string{
			".",
			"..",
			"some.thing",
			"some..thing",
			".name",
			"name.",
			"pass@word",
			"key#value",
			"data!info",
			"test$case",
			"name%test",
		}

		for _, input := range inputs {
			err := ValidateName(input)
			if !errors.Is(err, ErrInvalidCharacters) {
				t.Fatalf("ValidateName(%q): expected ErrInvalidCharacters, got %v", input, err)
			}
		}
	})

	t.Run("should not return error with valid names", func(t *testing.T) {
		inputs := []string{
			"pass-word",
			"mypassw0r3",
			"som3p4ssword_123",
			"VALID_NAME",
			"anotherValidName123",
		}

		for _, input := range inputs {
			err := ValidateName(input)
			if err != nil {
				t.Fatalf("ValidateName(%q): expected no error, got %v", input, err)
			}
		}
	})

}

func TestValidateWindowsReservedName(t *testing.T) {
	t.Run("should return error for Windows reserved names", func(t *testing.T) {
		inputs := []string{
			"con",
			"prn",
			"aux",
			"nul",
			"com1",
			"com2",
			"lpt1",
		}

		for _, input := range inputs {
			err := ValidateWindowsReservedName(input)
			if !errors.Is(err, ErrReservedName) {
				t.Fatalf("ValidateWindowsReservedName(%q): expected ErrReservedName, got %v", input, err)
			}
		}
	})
}
