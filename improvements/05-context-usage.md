# Group 5: Context Usage and Cancellation

**Priority:** 🟡 **MEDIUM** - Go best practices (Phase 2)

## Current State

The codebase passes context through layers, which is good! However, it's not fully utilized:

```go
// Current: Context is passed but not used for cancellation
func (s *Service) GetSecret(ctx context.Context, name string) ([]byte, error) {
    // ctx is never checked after initial validation
}
```

## Issues

1. **Context passed but not fully utilized**
2. **No timeout for potentially slow operations** (Argon2 key derivation)
3. **Context checked in storage layer** (good!) but not in service layer

## Learning Topics

- Context package in Go
- Cancellation patterns
- Timeouts and deadlines
- Context propagation
- When context is actually useful

## Understanding Context in Go

Context is primarily used for:

1. **Cancellation**: Stop work when the caller is no longer interested
2. **Timeouts**: Set deadlines for operations
3. **Request-scoped values**: Pass request IDs, user info, etc.

For a CLI application like MSK, context is less critical than for servers, but still useful for:
- Allowing Ctrl+C to cancel long operations
- Setting reasonable timeouts

## Implementation

### Option 1: Simple Timeout Wrapper (Recommended for CLI)

**Update: `internal/app/service.go`**

```go
package app

import (
    "context"
    "errors"
    "fmt"
    "time"
)

const (
    // DefaultOperationTimeout is the maximum time for a single operation
    DefaultOperationTimeout = 30 * time.Second
)

var ErrOperationTimeout = errors.New("operation timed out")

func (s *Service) AddSecret(ctx context.Context, name string, rawP []byte) error {
    // Add timeout for the entire operation
    ctx, cancel := context.WithTimeout(ctx, DefaultOperationTimeout)
    defer cancel()
    
    // Check for cancellation before starting
    if err := ctx.Err(); err != nil {
        return fmt.Errorf("operation cancelled: %w", err)
    }
    
    // Validate inputs
    if err := domain.ValidateSecretName(name); err != nil {
        return err
    }
    
    // Check cancellation before slow operation
    select {
    case <-ctx.Done():
        return fmt.Errorf("operation cancelled: %w", ctx.Err())
    default:
    }
    
    // Proceed with operation
    exists, err := s.repo.FileExists(ctx, name)
    if err != nil {
        return err
    }
    if exists {
        return ErrSecretExists
    }
    
    // ... encryption and saving (these use the ctx internally)
    
    return nil
}

func (s *Service) GetSecret(ctx context.Context, name string) ([]byte, error) {
    ctx, cancel := context.WithTimeout(ctx, DefaultOperationTimeout)
    defer cancel()
    
    if err := domain.ValidateSecretName(name); err != nil {
        return nil, err
    }
    
    // Check cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // ... rest of implementation
}
```

### Option 2: Cancellation-Aware Key Derivation

Note: Argon2 itself doesn't support cancellation, but we can check before and after:

**Update: `internal/encryption/encrypt.go`**

```go
func (a *ArgonCrypt) Encrypt(ctx context.Context, secret domain.Secret) (domain.EncryptedSecret, error) {
    // Check if already cancelled
    if err := ctx.Err(); err != nil {
        return domain.EncryptedSecret{}, fmt.Errorf("encryption cancelled: %w", err)
    }
    
    salt, err := randomBytes(MSK_SALT_SIZE)
    if err != nil {
        return domain.EncryptedSecret{}, err
    }
    
    mk := a.getMK()
    if mk == nil {
        return domain.EncryptedSecret{}, errors.New("master key not set")
    }
    
    // Note: Argon2 key derivation cannot be cancelled mid-operation
    // With 64MB memory and 2 iterations, this typically takes 200-500ms
    key := getArgonDeriveKey(mk, salt)
    defer secure.ZeroBytes(key)
    
    // Check if cancelled during key derivation
    if err := ctx.Err(); err != nil {
        return domain.EncryptedSecret{}, fmt.Errorf("encryption cancelled: %w", err)
    }
    
    // ... rest of encryption
}
```

### CLI Integration

**Update: `internal/cli/root.go`**

```go
import (
    "context"
    "os"
    "os/signal"
    "syscall"
)

func NewMSKCmd(service app.MSKService) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "msk",
        Short: "MSK Password Manager",
        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            // Setup context with signal handling for Ctrl+C
            ctx, cancel := context.WithCancel(cmd.Context())
            
            // Handle interrupt signals
            sigChan := make(chan os.Signal, 1)
            signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
            
            go func() {
                <-sigChan
                cancel()
            }()
            
            cmd.SetContext(ctx)
            
            // ... rest of pre-run (master key setup)
            return nil
        },
    }
    // ...
}
```

## When Context Matters Less

For a CLI password manager, context is less critical because:

1. **Operations are short** - Most operations complete in under a second
2. **No concurrent requests** - Only one user at a time
3. **Argon2 can't be cancelled** - The main slow operation isn't cancellable

However, using context properly is a good habit and prepares the code for:
- Integration tests with timeouts
- Potential future server mode
- Clean shutdown handling

## Pattern to Learn

**Context Usage Guidelines:**

1. **Accept context as first parameter** - Convention in Go
2. **Check context.Err() before slow operations** - Fail fast if cancelled
3. **Use context.WithTimeout for operations with known time limits**
4. **Pass context through all layers** - Don't break the chain
5. **Don't store context in structs** - Pass it as a parameter

**For CLI Applications:**
- Context is mostly useful for handling Ctrl+C gracefully
- Set reasonable timeouts to prevent hangs
- Don't over-engineer - simple cancellation checks are usually enough
