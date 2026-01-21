# Group 4: Input Validation and Sanitization

**Priority:** 🟡 **MEDIUM-HIGH** - Security and reliability (Phase 1)

## Current Problems

The application accepts any input for secret names without validation:

```bash
# These could cause problems:
msk add "../../../etc/passwd"    # Path traversal!
msk add "CON"                     # Reserved name on Windows
msk add ""                        # Empty name
msk add "$(rm -rf /)"             # Shell injection (if logged)
```

## Issues

1. **No input validation on secret names** - Path traversal risk with `../`, `/`, `\`
2. **No length limits** - Very long names could cause filesystem issues
3. **No character restrictions** - Special characters could break on some filesystems
4. **Windows reserved names not blocked** - CON, PRN, NUL, etc.

## Learning Topics

- Input validation principles
- Path traversal attacks
- Sanitization vs validation
- Filesystem security across platforms
- Defense in depth

## Implementation

### Step 1: Create Validation Package

**Create: `internal/domain/validation.go`**

```go
package domain

import (
    "errors"
    "regexp"
    "strings"
    "unicode/utf8"
)

var (
    ErrEmptyName         = errors.New("secret name cannot be empty")
    ErrNameTooLong       = errors.New("secret name too long (max 255 characters)")
    ErrInvalidCharacters = errors.New("secret name contains invalid characters")
    ErrPathTraversal     = errors.New("secret name cannot contain path separators")
    ErrReservedName      = errors.New("secret name is reserved by the operating system")
)

const (
    MaxSecretNameLength = 255
    MinSecretNameLength = 1
)

// validNameRegex allows alphanumeric, dash, underscore, dot, and spaces
// This is restrictive but safe across all platforms
var validNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._\- ]*$`)

// Windows reserved names that cannot be used as filenames
var windowsReservedNames = map[string]bool{
    "CON": true, "PRN": true, "AUX": true, "NUL": true,
    "COM1": true, "COM2": true, "COM3": true, "COM4": true,
    "COM5": true, "COM6": true, "COM7": true, "COM8": true, "COM9": true,
    "LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true,
    "LPT5": true, "LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
}

// ValidateSecretName validates a secret name for safety and filesystem compatibility.
// Returns nil if valid, or a descriptive error if invalid.
func ValidateSecretName(name string) error {
    // Check empty
    if name == "" {
        return ErrEmptyName
    }
    
    // Check length (in runes, not bytes, to handle unicode properly)
    runeCount := utf8.RuneCountInString(name)
    if runeCount > MaxSecretNameLength {
        return ErrNameTooLong
    }
    
    // Check for path traversal attempts
    if strings.Contains(name, "..") ||
        strings.Contains(name, "/") ||
        strings.Contains(name, "\\") ||
        strings.Contains(name, "\x00") {  // Null byte injection
        return ErrPathTraversal
    }
    
    // Check for valid characters
    if !validNameRegex.MatchString(name) {
        return ErrInvalidCharacters
    }
    
    // Check Windows reserved names (case-insensitive)
    upperName := strings.ToUpper(strings.TrimSpace(name))
    // Also check names with extensions like "CON.txt"
    if idx := strings.Index(upperName, "."); idx != -1 {
        upperName = upperName[:idx]
    }
    if windowsReservedNames[upperName] {
        return ErrReservedName
    }
    
    return nil
}

// ValidatePassword performs basic password validation.
// This is intentionally minimal - we don't want to be too restrictive.
func ValidatePassword(password []byte) error {
    if len(password) == 0 {
        return errors.New("password cannot be empty")
    }
    
    // Reasonable maximum to prevent memory issues
    if len(password) > 10*1024 { // 10KB max
        return errors.New("password too long (max 10KB)")
    }
    
    return nil
}
```

### Step 2: Add Tests for Validation

**Create: `internal/domain/validation_test.go`**

```go
package domain

import (
    "strings"
    "testing"
)

func TestValidateSecretName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr error
    }{
        // Valid names
        {"simple name", "github", nil},
        {"with dash", "my-password", nil},
        {"with underscore", "my_password", nil},
        {"with dot", "site.com", nil},
        {"with space", "my password", nil},
        {"with numbers", "github123", nil},
        
        // Invalid names
        {"empty", "", ErrEmptyName},
        {"path traversal dots", "../secret", ErrPathTraversal},
        {"path traversal slash", "foo/bar", ErrPathTraversal},
        {"path traversal backslash", "foo\\bar", ErrPathTraversal},
        {"null byte", "foo\x00bar", ErrPathTraversal},
        {"reserved CON", "CON", ErrReservedName},
        {"reserved con lowercase", "con", ErrReservedName},
        {"reserved with extension", "CON.txt", ErrReservedName},
        {"reserved NUL", "NUL", ErrReservedName},
        {"too long", strings.Repeat("a", 256), ErrNameTooLong},
        {"starts with dot", ".hidden", ErrInvalidCharacters},
        {"special chars", "pass<>word", ErrInvalidCharacters},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateSecretName(tt.input)
            if err != tt.wantErr {
                t.Errorf("ValidateSecretName(%q) = %v, want %v", tt.input, err, tt.wantErr)
            }
        })
    }
}

func TestValidatePassword(t *testing.T) {
    tests := []struct {
        name    string
        input   []byte
        wantErr bool
    }{
        {"valid password", []byte("mysecret"), false},
        {"empty password", []byte{}, true},
        {"very long password", make([]byte, 11*1024), true}, // 11KB
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidatePassword(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Step 3: Use Validation in Service Layer

**Update: `internal/app/service.go`**

```go
func (s *Service) AddSecret(ctx context.Context, name string, rawP []byte) error {
    // Validate inputs first
    if err := domain.ValidateSecretName(name); err != nil {
        return fmt.Errorf("invalid secret name: %w", err)
    }
    
    if err := domain.ValidatePassword(rawP); err != nil {
        return fmt.Errorf("invalid password: %w", err)
    }
    
    // Check if exists
    exists, err := s.repo.FileExists(ctx, name)
    if err != nil {
        return err
    }
    if exists {
        return ErrSecretExists
    }
    
    // ... rest of implementation
}

func (s *Service) GetSecret(ctx context.Context, name string) ([]byte, error) {
    if err := domain.ValidateSecretName(name); err != nil {
        return nil, fmt.Errorf("invalid secret name: %w", err)
    }
    
    // ... rest of implementation
}

func (s *Service) DeleteSecret(ctx context.Context, name string) error {
    if err := domain.ValidateSecretName(name); err != nil {
        return fmt.Errorf("invalid secret name: %w", err)
    }
    
    // ... rest of implementation
}
```

## Defense in Depth

Even with validation, the storage layer should also protect against path traversal:

**Update: `internal/storage/file/paths.go`**

```go
package file

import (
    "path/filepath"
    "strings"
)

// secretPath returns the filesystem path for a secret.
// It ensures the path stays within the vault directory.
func (s *Store) secretPath(name string) string {
    // Clean the path to remove any tricks
    cleanName := filepath.Clean(name)
    
    // Build the full path
    fullPath := filepath.Join(s.dir, cleanName+".msk")
    
    // Verify the path is still within the vault directory
    // This is defense in depth - validation should catch this first
    if !strings.HasPrefix(fullPath, filepath.Clean(s.dir)) {
        // This should never happen if validation is working
        // Return a safe path that will fail gracefully
        return filepath.Join(s.dir, "invalid")
    }
    
    return fullPath
}
```

## Pattern to Learn

**Input Validation Principles:**
- Validate at system boundaries (CLI, API)
- Use allowlists (valid characters) rather than blocklists
- Defense in depth - validate at multiple layers
- Fail securely - reject suspicious input
- Test validation with adversarial inputs
- Consider all target platforms (Windows reserved names!)
