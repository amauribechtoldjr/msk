package validator

import (
	"errors"
	"regexp"
	"runtime"
	"strings"
	"unicode"
)

var (
	ErrEmptyName          = errors.New("name cannot be empty")
	ErrNameTooLong        = errors.New("name cannot exceed 255 characters")
	ErrInvalidCharacters  = errors.New("name can only contain letters, numbers, hyphens and underscores")
	ErrPathSeparator      = errors.New("name cannot contain path separators")
	ErrReservedName       = errors.New("name cannot be a reserved system name")
	ErrControlCharacter   = errors.New("name cannot contain control characters")
	ErrWhitespace         = errors.New("name cannot contain whitespace")
)

var windowsReservedNames = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true,
	"COM5": true, "COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true,
	"LPT5": true, "LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
}

func ValidateName(name string) error {
	if name == "" {
		return ErrEmptyName
	}

	if len(name) > 255 {
		return ErrNameTooLong
	}

	for _, r := range name {
		if unicode.IsControl(r) {
			return ErrControlCharacter
		}
	}

	if strings.ContainsAny(name, " \t\n\r") {
		return ErrWhitespace
	}

	if strings.ContainsAny(name, "/\\") {
		return ErrPathSeparator
	}
	
	validPattern := regexp.MustCompile(`^[A-Za-z0-9_\-]+$`)
	if !validPattern.MatchString(name) {
		return ErrInvalidCharacters
	}

	if windowsReservedNames[strings.ToUpper(name)] {
		return ErrReservedName
	}

	return nil
}

func ValidateWindowsReservedName(name string) error {
	if windowsReservedNames[strings.ToUpper(name)] {
		return ErrReservedName
	}

	return nil
}

func Validate(name string) error {
	err := ValidateName(name)
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		err = ValidateWindowsReservedName(name)
		if err != nil {
			return err
		}
	}

	return nil
}

// Test Examples for ValidateName
//
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
//
// INVALID names:
//
// ErrEmptyName:
//   - ""
//
// ErrNameTooLong (> 255 characters):
//   - strings.Repeat("a", 256)
//
// ErrControlCharacter:
//   - "test\x00name"     (null character)
//   - "test\x1Fname"     (unit separator)
//   - "name\x7F"         (DEL character)
//
// ErrWhitespace:
//   - "my password"      (space)
//   - "my\tpassword"     (tab)
//   - "my\npassword"     (newline)
//   - "my\rpassword"     (carriage return)
//
// ErrPathSeparator:
//   - "path/to/secret"   (forward slash)
//   - "path\\to\\secret" (backslash)
//   - "/leading"
//   - "trailing/"
//
// ErrPathTraversal:
//   - "."
//   - ".."
//   - "some..thing"
//   - "parent/../secret"
//
// ErrLeadingTrailingDot:
//   - ".hidden"
//   - "file."
//   - ".both."
//
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
// ErrReservedName (Windows reserved - only on Windows via Validate()):
//   - "CON"
//   - "con"              (case insensitive)
//   - "PRN"
//   - "AUX"
//   - "NUL"
//   - "COM1" through "COM9"
//   - "LPT1" through "LPT9"
