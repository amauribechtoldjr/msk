# Architecture Improvements - Clean Architecture Design

**Priority:** 🟡 **MEDIUM** - Do this in Phase 2 (after error handling and input validation)

This document outlines architectural improvements to create a clean, maintainable foundation for the MSK password manager. These improvements focus on code organization, separation of concerns, and Go design patterns.

**Note:** Memory safety and secure memory management (SecureBuffer, memory locking) are covered separately in `01-memory-safety.md` and should be implemented last.

## Current Architecture Issues

1. **Service layer returns raw bytes** - Returns `[]byte` instead of domain types
2. **No clear layer boundaries** - Responsibilities are somewhat mixed
3. **Interface could be improved** - Service interface returns concrete types in some places
4. **Missing cleanup hooks** - No mechanism for cleanup after command execution

## Target Architecture: Clean Layered Design

### Core Principles

1. **Clear Boundaries**: Each layer has clear responsibilities
2. **Dependency Inversion**: Depend on abstractions, not concretions
3. **Single Responsibility**: Each component does one thing well
4. **Interface Segregation**: Small, focused interfaces

## Layer Responsibilities

```
┌─────────────────────────────────────────────────────┐
│  CLI Layer (cmd/, cli/)                             │
│  - Responsibility: User interaction, parsing args   │
│  - Owns: Command lifecycle, user prompts           │
│  - Depends on: Service interface                   │
│  - Handles: Clipboard operations (not service!)    │
└─────────────────────────────────────────────────────┘
                        ↓ (calls service methods)
┌─────────────────────────────────────────────────────┐
│  Service Layer (app/)                               │
│  - Responsibility: Business logic, orchestration   │
│  - Owns: Domain operations                         │
│  - Depends on: Repository & Encryption interfaces  │
│  - Returns: Domain types or errors                 │
└─────────────────────────────────────────────────────┘
                        ↓ (uses interfaces)
┌─────────────────────────────────────────────────────┐
│  Infrastructure Layer (storage/, encryption/)       │
│  - Responsibility: Technical implementation        │
│  - Owns: File I/O, cryptographic operations        │
│  - Implements: Repository & Encryption interfaces  │
└─────────────────────────────────────────────────────┘
```

## Implementation

### Step 1: Review Interface Design

**Current: `internal/app/service.go`**

The current interface is good, but consider if methods should return more specific types:

```go
type MSKService interface {
    AddSecret(ctx context.Context, name string, password []byte) error
    GetSecret(ctx context.Context, name string) ([]byte, error)  // Returns password bytes
    DeleteSecret(ctx context.Context, name string) error
    ListAll(ctx context.Context) error  // Should return []string, error
    ConfigMK(ctx context.Context, mk []byte)
}
```

**Improved version:**

```go
type MSKService interface {
    AddSecret(ctx context.Context, name string, password []byte) error
    GetSecret(ctx context.Context, name string) ([]byte, error)
    DeleteSecret(ctx context.Context, name string) error
    ListSecrets(ctx context.Context) ([]string, error)  // Better name and return type
    ConfigMK(ctx context.Context, mk []byte)
}
```

### Step 2: Ensure Proper Separation of Concerns

**Current issue:** The CLI layer correctly handles clipboard operations, which is good! Make sure this stays in the CLI layer.

**Good pattern (already in place):**

```go
// internal/cli/get.go - CLI handles clipboard
password, err := service.GetSecret(ctx, name)
if err != nil {
    return err
}
clip.Write(password)  // Clipboard is CLI responsibility, not service
```

**Bad pattern (avoid):**

```go
// DON'T put clipboard logic in service layer
func (s *Service) GetSecret(...) {
    // ... decrypt ...
    clip.Write(password)  // ❌ Service shouldn't know about clipboard
    return password
}
```

### Step 3: Add Command Lifecycle Hooks

Add `PersistentPostRun` for cleanup operations. This is useful for future memory safety improvements.

**Update: `internal/cli/root.go`**

```go
func NewMSKCmd(service app.MSKService) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "msk",
        Short: "MSK Password Manager",
        Long:  `MSK is a lightweight password manager...`,
        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()
            mk, err := PromptPassword("Enter master password:")
            if err != nil {
                return err
            }
            service.ConfigMK(ctx, mk)
            return nil
        },
        PersistentPostRun: func(cmd *cobra.Command, args []string) {
            // Cleanup hook - useful for future memory clearing
            // For now, this is a placeholder for cleanup operations
        },
    }
    // ... add subcommands ...
    return cmd
}
```

### Step 4: Improve Repository Interface

**Current: `internal/storage/repository.go`**

```go
type Repository interface {
    SaveFile(ctx context.Context, encryption domain.EncryptedSecret, name string) error
    GetFile(ctx context.Context, name string) ([]byte, error)
    DeleteFile(ctx context.Context, name string) (bool, error)
    FileExists(ctx context.Context, name string) (bool, error)
}
```

**Add ListFiles method:**

```go
type Repository interface {
    SaveFile(ctx context.Context, encryption domain.EncryptedSecret, name string) error
    GetFile(ctx context.Context, name string) ([]byte, error)
    DeleteFile(ctx context.Context, name string) (bool, error)
    FileExists(ctx context.Context, name string) (bool, error)
    ListFiles(ctx context.Context) ([]string, error)  // New method
}
```

### Step 5: Keep Domain Types Clean

**Current: `internal/domain/secret.go`**

Keep domain types simple and focused:

```go
package domain

import "time"

// Secret represents a stored password entry
type Secret struct {
    Name      string    `json:"name"`
    Password  []byte    `json:"password"`
    CreatedAt time.Time `json:"created_at"`
}

// EncryptedSecret represents encrypted data ready for storage
type EncryptedSecret struct {
    Data  []byte
    Salt  [16]byte
    Nonce [12]byte
}
```

## Design Patterns to Follow

### 1. Dependency Injection

Pass dependencies through constructors, not globals:

```go
// Good: Dependencies passed in
func NewMSKService(r storage.Repository, c encryption.Encryption) *Service {
    return &Service{repo: r, crypto: c}
}

// Bad: Global dependencies
var globalRepo storage.Repository
func NewService() *Service {
    return &Service{repo: globalRepo}
}
```

### 2. Interface-Based Design

Depend on interfaces, not concrete types:

```go
// Good: Service depends on interface
type Service struct {
    repo   storage.Repository  // Interface
    crypto encryption.Encryption  // Interface
}

// This allows easy testing with mocks
```

### 3. Error Wrapping

Wrap errors with context:

```go
// Good: Wrapped error with context
if err != nil {
    return fmt.Errorf("failed to save secret %s: %w", name, err)
}

// Bad: Raw error without context
if err != nil {
    return err
}
```

## Benefits of Clean Architecture

1. **Testability**: Each layer can be tested independently with mocks
2. **Maintainability**: Changes in one layer don't affect others
3. **Flexibility**: Easy to swap implementations (e.g., different storage backends)
4. **Readability**: Clear separation makes code easier to understand

## What This Document Does NOT Cover

The following topics are covered in `01-memory-safety.md` (Phase 5):
- SecureBuffer implementation
- Memory locking (mlock/VirtualLock)
- Secure memory zeroing
- Password clearing from memory

Focus on getting the architecture right first. Memory safety can be added later without major refactoring if the architecture is clean.

---

*Clean architecture makes it easier to add security features later. Build the foundation first.*
