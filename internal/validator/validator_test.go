package validator

import (
	"runtime"
	"strings"
	"testing"
)

type TestCase struct {
	name    string
	input   []string
	wantErr error
}

func runTestCases(t *testing.T, tests []TestCase, validateFn func(string) error) {
	for _, tt := range tests {
		for _, input := range tt.input {
			t.Run(tt.name, func(t *testing.T) {
				result := validateFn(input)

				if result != tt.wantErr {
					t.Logf("Test failed: %v", tt.name)
					t.Logf("ValidateName('%s') = %v", input, result)
					t.Errorf("Expects: %v", tt.wantErr)
				}
			})
		}
	}
}

func TestValidateName(t *testing.T) {
	tests := []TestCase{
		{
			name:    "should return error for empty name",
			input:   []string{""},
			wantErr: ErrEmptyName,
		},
		{
			name:    "should return error for too long name",
			input:   []string{strings.Repeat("T", 256)},
			wantErr: ErrNameTooLong,
		},
		{
			name: "should return error for name with null character",
			input: []string{
				"test\x00name",
				"test\x1Fname",
				"test\x7Fname",
				"my\tpassword",
				"my\npassword",
				"my\rpassword",
			},
			wantErr: ErrControlCharacter,
		},
		{
			name:    "should return error for name with space",
			input:   []string{"my password"},
			wantErr: ErrWhitespace,
		},
		{
			name: "should return error for name with path separators",
			input: []string{
				"path/to/secret",
				"path\\to\\secret",
				"trailing/",
				"/leading",
			},
			wantErr: ErrPathSeparator,
		},
		{
			name: "should return error for name with invalid characters",
			input: []string{
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
			},
			wantErr: ErrInvalidCharacters,
		},
		{
			name: "should not return error with valid names",
			input: []string{
				"pass-word",
				"mypassw0r3",
				"som3p4ssword_123",
				"VALID_NAME",
				"anotherValidName123",
			},
			wantErr: nil,
		},
	}

	runTestCases(t, tests, ValidateName)
}

func TestValidateWindowsReservedName(t *testing.T) {

	if runtime.GOOS == "windows" {
		tests := []TestCase{
			{
				name: "should return error for Windows reserved names",
				input: []string{
					"con",
					"prn",
					"aux",
					"nul",
					"com1",
					"com2",
					"lpt1",
				},
				wantErr: ErrReservedName,
			},
		}

		runTestCases(t, tests, ValidateWindowsReservedName)
	}
}
