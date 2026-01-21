# Group 11: Missing Features and Incomplete Implementation

**Priority:** 🟡 **MEDIUM** - Feature completeness (Phase 1)

## Current Problems

### Problem 1: ListAll() Not Implemented

**File: `internal/app/service.go`**

```go
func (s *Service) ListAll(ctx context.Context) error {
    return nil  // Does nothing!
}
```

### Problem 2: No Unit Tests

The entire codebase has no test files.

### Problem 3: Incorrect Success Message

**File: `internal/cli/get.go` (potential issue)**

Some commands might have copy-paste errors in success messages.

## Issues

1. **`ListAll()` not implemented** - Returns nil, does nothing
2. **Missing unit tests** - No test coverage at all
3. **Return type wrong** - `ListAll` should return `[]string, error`

## Learning Topics

- Feature implementation
- Unit testing in Go
- Table-driven tests
- Mocking with interfaces
- Test coverage

## Implementation

### Step 1: Implement ListAll (ListSecrets)

**Update: `internal/storage/repository.go`**

```go
package storage

import (
    "context"
    
    "github.com/amauribechtoldjr/msk/internal/domain"
)

type Repository interface {
    SaveFile(ctx context.Context, data domain.EncryptedSecret, name string) error
    GetFile(ctx context.Context, name string) ([]byte, error)
    DeleteFile(ctx context.Context, name string) (bool, error)
    FileExists(ctx context.Context, name string) (bool, error)
    ListFiles(ctx context.Context) ([]string, error)  // NEW
}
```

**Update: `internal/storage/file/store.go`**

```go
import (
    "os"
    "strings"
)

// ListFiles returns all secret names in the vault (without decrypting)
func (s *Store) ListFiles(ctx context.Context) ([]string, error) {
    // Check context
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Read vault directory
    entries, err := os.ReadDir(s.dir)
    if err != nil {
        if os.IsNotExist(err) {
            return []string{}, nil  // Empty vault
        }
        return nil, fmt.Errorf("failed to read vault directory: %w", err)
    }
    
    var names []string
    for _, entry := range entries {
        // Skip directories
        if entry.IsDir() {
            continue
        }
        
        // Only include .msk files
        name := entry.Name()
        if strings.HasSuffix(name, ".msk") {
            // Remove extension to get secret name
            secretName := strings.TrimSuffix(name, ".msk")
            names = append(names, secretName)
        }
    }
    
    return names, nil
}
```

**Update: `internal/app/service.go`**

```go
// ListSecrets returns all stored secret names (without passwords)
func (s *Service) ListSecrets(ctx context.Context) ([]string, error) {
    names, err := s.repo.ListFiles(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to list secrets: %w", err)
    }
    return names, nil
}
```

### Step 2: Create List CLI Command

**Create: `internal/cli/list.go`**

```go
package cli

import (
    "fmt"
    "sort"

    "github.com/amauribechtoldjr/msk/internal/app"
    "github.com/spf13/cobra"
)

func NewListCmd(service app.MSKService) *cobra.Command {
    cmd := &cobra.Command{
        Use:     "list",
        Aliases: []string{"ls", "l"},
        Short:   "List all stored password names",
        Long: `List all password entries in the vault.

Only the names are displayed, not the actual passwords.
Use 'msk get <name>' to retrieve a specific password.`,
        Example: `  msk list
  msk ls`,
        Args:          cobra.NoArgs,
        SilenceErrors: true,
        SilenceUsage:  true,
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()

            names, err := service.ListSecrets(ctx)
            if err != nil {
                return fmt.Errorf("failed to list secrets: %w", err)
            }

            if len(names) == 0 {
                fmt.Println("No passwords stored yet.")
                fmt.Println("\nUse 'msk add <name>' to add your first password.")
                return nil
            }

            // Sort alphabetically for consistent output
            sort.Strings(names)

            fmt.Printf("Stored passwords (%d):\n\n", len(names))
            for _, name := range names {
                fmt.Printf("  • %s\n", name)
            }

            return nil
        },
    }

    return cmd
}
```

**Update: `internal/cli/root.go`** - Add the list command

```go
cmd.AddCommand(NewListCmd(service))
```

### Step 3: Add Unit Tests

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
        wantErr bool
    }{
        // Valid cases
        {"simple", "github", false},
        {"with dash", "my-password", false},
        {"with underscore", "my_password", false},
        {"with numbers", "password123", false},
        {"with dot", "site.com", false},
        
        // Invalid cases
        {"empty", "", true},
        {"path traversal", "../secret", true},
        {"forward slash", "foo/bar", true},
        {"backslash", "foo\\bar", true},
        {"too long", strings.Repeat("a", 256), true},
        {"reserved CON", "CON", true},
        {"reserved NUL", "NUL", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateSecretName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateSecretName(%q) error = %v, wantErr %v", 
                    tt.input, err, tt.wantErr)
            }
        })
    }
}
```

**Create: `internal/app/service_test.go`**

```go
package app

import (
    "context"
    "testing"
    
    "github.com/amauribechtoldjr/msk/internal/domain"
)

// MockRepository for testing
type MockRepository struct {
    files    map[string][]byte
    saveErr  error
    getErr   error
}

func NewMockRepository() *MockRepository {
    return &MockRepository{
        files: make(map[string][]byte),
    }
}

func (m *MockRepository) SaveFile(ctx context.Context, data domain.EncryptedSecret, name string) error {
    if m.saveErr != nil {
        return m.saveErr
    }
    m.files[name] = data.Data
    return nil
}

func (m *MockRepository) GetFile(ctx context.Context, name string) ([]byte, error) {
    if m.getErr != nil {
        return nil, m.getErr
    }
    data, ok := m.files[name]
    if !ok {
        return nil, ErrSecretNotFound
    }
    return data, nil
}

func (m *MockRepository) DeleteFile(ctx context.Context, name string) (bool, error) {
    if _, ok := m.files[name]; !ok {
        return false, ErrSecretNotFound
    }
    delete(m.files, name)
    return true, nil
}

func (m *MockRepository) FileExists(ctx context.Context, name string) (bool, error) {
    _, ok := m.files[name]
    return ok, nil
}

func (m *MockRepository) ListFiles(ctx context.Context) ([]string, error) {
    var names []string
    for name := range m.files {
        names = append(names, name)
    }
    return names, nil
}

// MockEncryption for testing
type MockEncryption struct {
    mk []byte
}

func (m *MockEncryption) Encrypt(secret domain.Secret) (domain.EncryptedSecret, error) {
    return domain.EncryptedSecret{
        Data: []byte("encrypted:" + string(secret.Password)),
    }, nil
}

func (m *MockEncryption) Decrypt(data []byte) (domain.Secret, error) {
    return domain.Secret{
        Password: []byte("decrypted"),
    }, nil
}

func (m *MockEncryption) ConfigMK(mk []byte) {
    m.mk = mk
}

// Tests
func TestService_AddSecret(t *testing.T) {
    repo := NewMockRepository()
    enc := &MockEncryption{}
    svc := NewMSKService(repo, enc)
    svc.ConfigMK(context.Background(), []byte("masterkey"))
    
    ctx := context.Background()
    
    t.Run("add new secret", func(t *testing.T) {
        err := svc.AddSecret(ctx, "github", []byte("password123"))
        if err != nil {
            t.Errorf("AddSecret() error = %v", err)
        }
    })
    
    t.Run("add duplicate secret", func(t *testing.T) {
        err := svc.AddSecret(ctx, "github", []byte("password456"))
        if err != ErrSecretExists {
            t.Errorf("AddSecret() expected ErrSecretExists, got %v", err)
        }
    })
    
    t.Run("invalid name", func(t *testing.T) {
        err := svc.AddSecret(ctx, "../invalid", []byte("password"))
        if err == nil {
            t.Error("AddSecret() expected error for invalid name")
        }
    })
}

func TestService_GetSecret(t *testing.T) {
    repo := NewMockRepository()
    enc := &MockEncryption{}
    svc := NewMSKService(repo, enc)
    svc.ConfigMK(context.Background(), []byte("masterkey"))
    
    ctx := context.Background()
    
    // Add a secret first
    _ = svc.AddSecret(ctx, "test", []byte("password"))
    
    t.Run("get existing secret", func(t *testing.T) {
        password, err := svc.GetSecret(ctx, "test")
        if err != nil {
            t.Errorf("GetSecret() error = %v", err)
        }
        if password == nil {
            t.Error("GetSecret() returned nil password")
        }
    })
    
    t.Run("get non-existent secret", func(t *testing.T) {
        _, err := svc.GetSecret(ctx, "nonexistent")
        if err == nil {
            t.Error("GetSecret() expected error for non-existent secret")
        }
    })
}

func TestService_ListSecrets(t *testing.T) {
    repo := NewMockRepository()
    enc := &MockEncryption{}
    svc := NewMSKService(repo, enc)
    svc.ConfigMK(context.Background(), []byte("masterkey"))
    
    ctx := context.Background()
    
    t.Run("empty vault", func(t *testing.T) {
        names, err := svc.ListSecrets(ctx)
        if err != nil {
            t.Errorf("ListSecrets() error = %v", err)
        }
        if len(names) != 0 {
            t.Errorf("ListSecrets() expected empty, got %d items", len(names))
        }
    })
    
    t.Run("with secrets", func(t *testing.T) {
        _ = svc.AddSecret(ctx, "github", []byte("pass1"))
        _ = svc.AddSecret(ctx, "email", []byte("pass2"))
        
        names, err := svc.ListSecrets(ctx)
        if err != nil {
            t.Errorf("ListSecrets() error = %v", err)
        }
        if len(names) != 2 {
            t.Errorf("ListSecrets() expected 2, got %d", len(names))
        }
    })
}
```

### Step 4: Run Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Test Organization

```
internal/
├── app/
│   ├── service.go
│   └── service_test.go      # Service tests with mocks
├── domain/
│   ├── secret.go
│   ├── validation.go
│   └── validation_test.go   # Validation tests
├── storage/
│   └── file/
│       ├── store.go
│       └── store_test.go    # Integration tests (use temp dirs)
└── encryption/
    ├── encrypt.go
    ├── decrypt.go
    └── encryption_test.go   # Crypto tests
```

## Pattern to Learn

**Go Testing Best Practices:**
- Use table-driven tests for multiple cases
- Create mock implementations for interfaces
- Test both success and error paths
- Use `t.Run()` for subtests
- Keep tests close to the code they test (`_test.go` suffix)

**Test Coverage:**
- Aim for 70-80% coverage on critical paths
- Don't chase 100% - focus on important logic
- Use `go test -cover` to check coverage
