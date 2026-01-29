package validator

import (
	"strings"
	"testing"
)

type TestCase struct {
	name    string
	input   string
	wantErr error
}

func runTestCases(t *testing.T, tests []TestCase) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateName(tt.input)

			if result != tt.wantErr {
				t.Logf("Test failed: %v", tt.name)
				t.Logf("ValidateName() = %v", result)
				t.Errorf("Expects: %v", tt.wantErr)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []TestCase{
		{
			name: "should return error for empty name",
			input: "",
			wantErr: ErrEmptyName,
		},
		{
			name: "should return error for too long name",
			input: strings.Repeat("T", 256),
			wantErr: ErrNameTooLong,
		},
		{
			name: "should return error for name with null character",
			input: "test\x00name",
			wantErr: ErrControlCharacter,
		},
		{
			name: "should return error for name with unit separator",
			input: "test\x1Fname",
			wantErr: ErrControlCharacter,
		},
		{
			name: "should return error for name with DEL character",
			input: "test\x7Fname",
			wantErr: ErrControlCharacter,
		},
		{
			name: "should return error for name with space",
			input: "my password",
			wantErr: ErrWhitespace,
		},
		{
			name: "should return error for name with tab",
			input: "my\tpassword",
			wantErr: ErrControlCharacter,	
		},
		{
			name: "should return error for name with newline",
			input: "my\npassword",
			wantErr: ErrControlCharacter,
		},
		{
			name: "should return error for name with carriage return",
			input: "my\rpassword",
			wantErr: ErrControlCharacter,
		},
		{
			name: "should return error for name with path separator (forward slash)",
			input: "path/to/secret",
			wantErr: ErrPathSeparator,
		},
		{
			name: "should return error for name with path separator (backslash)",
			input: "path\\to\\secret",
			wantErr: ErrPathSeparator,
		},
		{
			name: "should return error for name with path separator (trailing slash)",
			input: "trailing/",
			wantErr: ErrPathSeparator,
		},
		{
			name: "should return error for name with path separator (leading slash)",
			input: "/leading",
			wantErr: ErrPathSeparator,
		},
		{
			name: "should return error for name with invalid characters (dot only)",
			input: ".",
			wantErr: ErrInvalidCharacters,
		},
		{
			name: "should return error for name with invalid characters (double dot only)",
			input: "..",
			wantErr: ErrInvalidCharacters,
		},
		{
			name: "should return error for name with invalid characters (string with single dots)",
			input: "some.thing",
			wantErr: ErrInvalidCharacters,
		},
		{
			name: "should return error for name with invalid characters (string with double dots)",
			input: "some..thing",
			wantErr: ErrInvalidCharacters,
		},
		{
			name: "should return error for name with invalid characters (leading dot)",
			input: ".name",
			wantErr: ErrInvalidCharacters,
		},
		{
			name: "should return error for name with invalid characters (trailing dot)",
			input: "name.",
			wantErr: ErrInvalidCharacters,
		},
	}

	runTestCases(t, tests)
}
// VALID names (should return nil):
//   - "my-password"
//   - "MyPassword123"
//   - "secret_key"
//   - "API-KEY-2024"
//   - "a"
//   - "A1_b2-C3"
//   - "ALLCAPS"
//   - "lowercase"
//   - "with-hyphen"
//   - "with_underscore"

// ErrInvalidCharacters:
//   - "my@password"      (@ symbol)
//   - "pass#word"        (hash)
//   - "secret!"          (exclamation)
//   - "test$var"         (dollar sign)
//   - "name%value"       (percent)
//   - "key=value"        (equals)
//   - "hello+world"      (plus)
//   - "file.txt"         (dot in middle - also caught by dot rules)
//   - "日本語"            (unicode characters)
//
// ErrReservedName (Windows reserved - only for windows Users):
//   - "CON"
//   - "con"              (case insensitive)
//   - "PRN"
//   - "AUX"
//   - "NUL"
//   - "COM1" through "COM9"
//   - "LPT1" through "LPT9"