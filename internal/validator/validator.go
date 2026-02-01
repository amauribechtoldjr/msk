package validator

import (
	"errors"
	"regexp"
	"runtime"
	"strings"
	"unicode"
)

var (
	ErrEmptyName         = errors.New("name cannot be empty")
	ErrNameTooLong       = errors.New("name cannot exceed 255 characters")
	ErrInvalidCharacters = errors.New("name can only contain letters, numbers, hyphens and underscores")
	ErrPathSeparator     = errors.New("name cannot contain path separators")
	ErrReservedName      = errors.New("name cannot be a reserved system name")
	ErrControlCharacter  = errors.New("name cannot contain control characters")
	ErrWhitespace        = errors.New("name cannot contain whitespace")
)

var windowsReservedNames = map[string]bool{
	"con": true, "prn": true, "aux": true, "nul": true,
	"com1": true, "com2": true, "com3": true, "com4": true,
	"com5": true, "com6": true, "com7": true, "com8": true, "com9": true,
	"lpt1": true, "lpt2": true, "lpt3": true, "lpt4": true,
	"lpt5": true, "lpt6": true, "lpt7": true, "lpt8": true, "lpt9": true,
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

	return nil
}

func ValidateWindowsReservedName(name string) error {
	if windowsReservedNames[strings.ToLower(name)] {
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
