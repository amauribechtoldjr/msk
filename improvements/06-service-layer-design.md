# Group 6: Service Layer Design and Separation of Concerns

**Priority:** 🟡 **MEDIUM** - Architecture and maintainability (Phase 2)

## Current State

The current service layer is actually well-designed in many ways:
- Uses interfaces for dependencies (Repository, Encryption)
- Returns domain types or raw data appropriately
- Clipboard handling is in CLI layer (correct!)

However, there are some improvements to consider.

## Issues

1. **ListAll() returns error only** - Should return `[]string, error`
2. **Service interface could be cleaner** - Consider what each method returns
3. **No explicit cleanup mechanism** - For future memory safety integration

## What's Already Good

The codebase correctly separates concerns:

```go
// CLI layer handles clipboard - CORRECT!
password, err := service.GetSecret(ctx, name)
if err != nil {
    return err
}
clip.Write(password)  // CLI responsibility
```

This is the right approach! The service layer should not know about clipboard operations.

## Implementation

### Step 1: Improve Service Interface

**Update: `internal/app/service.go`**

```go
package app

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/amauribechtoldjr/msk/internal/domain"
    "github.com/amauribechtoldjr/msk/internal/encryption"
    "github.com/amauribechtoldjr/msk/internal/storage"
)

var (
    ErrSecretExists   = errors.New("secret already exists")
    ErrSecretNotFound = errors.New("secret not found")
)

// MSKService defines the contract for password management operations.
// This interface allows for easy mocking in tests.
type MSKService interface {
    // AddSecret stores a new password with the given name
    AddSecret(ctx context.Context, name string, password []byte) error
    
    // GetSecret retrieves the password for the given name
    GetSecret(ctx context.Context, name string) ([]byte, error)
    
    // DeleteSecret removes the password with the given name
    DeleteSecret(ctx context.Context, name string) error
    
    // ListSecrets returns all stored secret names (not the passwords)
    ListSecrets(ctx context.Context) ([]string, error)
    
    // ConfigMK sets the master key for encryption operations
    ConfigMK(ctx context.Context, mk []byte)
    
    // Close performs cleanup operations (for future memory clearing)
    Close() error
}

// Service implements MSKService
type Service struct {
    repo   storage.Repository
    crypto encryption.Encryption
}

// Verify Service implements MSKService at compile time
var _ MSKService = (*Service)(nil)

// NewMSKService creates a new service with the given dependencies
func NewMSKService(repo storage.Repository, crypto encryption.Encryption) *Service {
    return &Service{
        repo:   repo,
        crypto: crypto,
    }
}

func (s *Service) ConfigMK(ctx context.Context, mk []byte) {
    s.crypto.ConfigMK(mk)
}

func (s *Service) Close() error {
    // Future: Clear master key from memory
    // s.crypto.ClearMK()
    return nil
}

func (s *Service) AddSecret(ctx context.Context, name string, rawP []byte) error {
    // Validation
    if err := domain.ValidateSecretName(name); err != nil {
        return fmt.Errorf("invalid name: %w", err)
    }
    
    // Check existence
    exists, err := s.repo.FileExists(ctx, name)
    if err != nil {
        return fmt.Errorf("failed to check existence: %w", err)
    }
    if exists {
        return ErrSecretExists
    }
    
    // Create domain object
    secret := domain.Secret{
        Name:      name,
        Password:  rawP,
        CreatedAt: time.Now().UTC(),
    }
    
    // Encrypt
    encrypted, err := s.crypto.Encrypt(secret)
    if err != nil {
        return fmt.Errorf("encryption failed: %w", err)
    }
    
    // Store
    if err := s.repo.SaveFile(ctx, encrypted, name); err != nil {
        return fmt.Errorf("failed to save: %w", err)
    }
    
    return nil
}

func (s *Service) GetSecret(ctx context.Context, name string) ([]byte, error) {
    // Validation
    if err := domain.ValidateSecretName(name); err != nil {
        return nil, fmt.Errorf("invalid name: %w", err)
    }
    
    // Check existence
    exists, err := s.repo.FileExists(ctx, name)
    if err != nil {
        return nil, fmt.Errorf("failed to check existence: %w", err)
    }
    if !exists {
        return nil, ErrSecretNotFound
    }
    
    // Read encrypted data
    fileData, err := s.repo.GetFile(ctx, name)
    if err != nil {
        return nil, fmt.Errorf("failed to read: %w", err)
    }
    
    // Decrypt
    secret, err := s.crypto.Decrypt(fileData)
    if err != nil {
        return nil, fmt.Errorf("decryption failed: %w", err)
    }
    
    return secret.Password, nil
}

func (s *Service) DeleteSecret(ctx context.Context, name string) error {
    // Validation
    if err := domain.ValidateSecretName(name); err != nil {
        return fmt.Errorf("invalid name: %w", err)
    }
    
    _, err := s.repo.DeleteFile(ctx, name)
    if err != nil {
        return fmt.Errorf("failed to delete: %w", err)
    }
    
    return nil
}

func (s *Service) ListSecrets(ctx context.Context) ([]string, error) {
    return s.repo.ListFiles(ctx)
}
```

### Step 2: Update Repository Interface

**Update: `internal/storage/repository.go`**

```go
package storage

import (
    "context"
    
    "github.com/amauribechtoldjr/msk/internal/domain"
)

// Repository defines the contract for secret storage operations.
type Repository interface {
    // SaveFile saves an encrypted secret with the given name
    SaveFile(ctx context.Context, data domain.EncryptedSecret, name string) error
    
    // GetFile retrieves the encrypted data for a secret
    GetFile(ctx context.Context, name string) ([]byte, error)
    
    // DeleteFile removes a secret by name
    DeleteFile(ctx context.Context, name string) (bool, error)
    
    // FileExists checks if a secret with the given name exists
    FileExists(ctx context.Context, name string) (bool, error)
    
    // ListFiles returns all secret names (without decrypting)
    ListFiles(ctx context.Context) ([]string, error)
}
```

### Step 3: Implement ListFiles in Store

**Update: `internal/storage/file/store.go`**

```go
func (s *Store) ListFiles(ctx context.Context) ([]string, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    entries, err := os.ReadDir(s.dir)
    if err != nil {
        return nil, fmt.Errorf("failed to read vault directory: %w", err)
    }
    
    var names []string
    for _, entry := range entries {
        if entry.IsDir() {
            continue
        }
        
        name := entry.Name()
        // Only include .msk files
        if strings.HasSuffix(name, ".msk") {
            // Remove the .msk extension
            names = append(names, strings.TrimSuffix(name, ".msk"))
        }
    }
    
    return names, nil
}
```

## Layer Responsibilities Summary

```
┌─────────────────────────────────────────────────┐
│  CLI Layer (internal/cli/)                       │
│  - Parse command line arguments                  │
│  - Prompt for user input (master key, passwords)│
│  - Display output and errors                     │
│  - Handle clipboard operations                   │
│  - Convert user input to domain calls            │
└─────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────┐
│  Service Layer (internal/app/)                   │
│  - Validate business rules                       │
│  - Coordinate between encryption and storage     │
│  - Handle domain errors                          │
│  - NO knowledge of CLI or clipboard              │
└─────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────┐
│  Domain Layer (internal/domain/)                 │
│  - Define data structures (Secret, Encrypted)    │
│  - Define validation rules                       │
│  - Define domain errors                          │
│  - NO dependencies on other layers               │
└─────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────┐
│  Infrastructure (encryption/, storage/)          │
│  - Implement technical details                   │
│  - File I/O, cryptographic operations            │
│  - Implement repository interface                │
└─────────────────────────────────────────────────┘
```

## Pattern to Learn

**Clean Architecture Principles:**
- **Dependency Rule**: Inner layers don't depend on outer layers
- **Interface Segregation**: Small, focused interfaces
- **Single Responsibility**: Each layer has one job
- **Testability**: Interfaces allow mocking in tests

**Go-Specific Patterns:**
- Use `var _ Interface = (*Struct)(nil)` to verify interface compliance
- Return `error` as the last return value
- Accept interfaces, return structs
- Keep interfaces small (often 1-3 methods)
