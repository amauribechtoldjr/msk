# MSK Code Review - Improvement Plan

This document contains a comprehensive code review summary, organized by learning topics with security and Go design patterns as priorities.

## Table of Contents

1. [Group 1: Critical Security - Memory Safety](#group-1-critical-security---memory-safety--sensitive-data-handling)
2. [Group 2: Cryptographic Security - Argon2 Configuration](#group-2-cryptographic-security---argon2-configuration)
3. [Group 3: Go Error Handling & Defensive Programming](#group-3-go-error-handling--defensive-programming)
4. [Group 4: Input Validation & Sanitization](#group-4-input-validation--sanitization)
5. [Group 5: Context Usage & Cancellation](#group-5-context-usage--cancellation)
6. [Group 6: Service Layer Design & Separation of Concerns](#group-6-service-layer-design--separation-of-concerns)
7. [Group 7: Storage & File System Security](#group-7-storage--file-system-security)
8. [Group 8: Dependency Management & Module Configuration](#group-8-dependency-management--module-configuration)
9. [Group 9: Cobra CLI Best Practices](#group-9-cobra-cli-best-practices)
10. [Group 10: Configuration & Flexibility](#group-10-configuration--flexibility)
11. [Group 11: Missing Features & Incomplete Implementation](#group-11-missing-features--incomplete-implementation)

---

## Group 1: Critical Security - Memory Safety & Sensitive Data Handling

**Priority:** 🔴 **CRITICAL** - Do this first!

### Issues

1. **Master key never cleared from memory** (`internal/encryption/encryption.go:23`)
2. **Passwords stored in memory but not cleared after use**
3. **Derived encryption keys not cleared after encryption** (only after decryption)
4. **Password output security** (✅ Already partially fixed - now using clipboard)

### Learning Topics

- Secure memory management in Go
- Preventing sensitive data exposure in memory dumps
- Using `runtime.KeepAlive()` to prevent compiler optimizations
- Zeroing byte slices securely

### Implementation Examples

**Note on `runtime.KeepAlive`:** While `runtime.KeepAlive()` can prevent compiler optimizations, using it frequently is often a code smell indicating design issues. We'll explore better alternatives below.

#### Option A: Simple Zeroing (Recommended for Most Cases)

**Create: `internal/encryption/secure.go`**

```go
package encryption

// ZeroBytes zeroes out a byte slice. 
// For local variables and function parameters, this is sufficient as 
// the compiler rarely optimizes away writes to parameters/return values.
func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
```

**Why this works:**
- For local variables and function parameters, compiler optimizations that remove writes are rare
- Simpler code, easier to understand
- Sufficient for most practical security purposes
- No need for `runtime.KeepAlive` tricks

#### Option B: Architecture-Based Approach (Best Long-term)

Instead of fighting the compiler, **minimize the need for explicit clearing** by using references and single ownership:

**Strategy: Single Source of Truth with Reference Passing**

```go
package encryption

import (
	"sync"
	"github.com/amauribechtoldjr/msk/internal/domain"
)

type ArgonCrypt struct {
	mk []byte
	mu sync.RWMutex
}

// ConfigMK takes ownership of the master key - caller should not modify after this call
// The ArgonCrypt struct owns the memory and is responsible for clearing it
func (ac *ArgonCrypt) ConfigMK(mk []byte) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	
	// Clear old master key first
	if ac.mk != nil {
		ZeroBytes(ac.mk)
		ac.mk = nil
	}
	
	// Take ownership - make a copy that we control
	ac.mk = make([]byte, len(mk))
	copy(ac.mk, mk)
}

// ClearMK clears the master key - should be called after operations complete
func (ac *ArgonCrypt) ClearMK() {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if ac.mk != nil {
		ZeroBytes(ac.mk)
		ac.mk = nil
	}
}

// getMK returns a reference to the master key (internal use only)
// Operations use this reference, avoiding copies
func (ac *ArgonCrypt) getMK() []byte {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.mk  // Returns reference - still owned by ArgonCrypt
}
```

**Benefits:**
- Single copy of master key in memory (in ArgonCrypt)
- Operations use reference, no unnecessary copies
- Clear ownership model - ArgonCrypt owns it, clears it
- No need for `runtime.KeepAlive` because we're working with owned memory

#### Option C: Using Memguard Library (Professional Solution)

If you want professional-grade secure memory handling, consider the `memguard` library:

```bash
go get github.com/awnumar/memguard
```

**Example usage:**

```go
package encryption

import (
	"github.com/awnumar/memguard"
	"github.com/amauribechtoldjr/msk/internal/domain"
)

type ArgonCrypt struct {
	mk *memguard.Enclave  // Secure memory container
}

func NewArgonCrypt() *ArgonCrypt {
	return &ArgonCrypt{}
}

func (ac *ArgonCrypt) ConfigMK(mk []byte) error {
	// Clear old key
	if ac.mk != nil {
		ac.mk.Destroy()
	}
	
	// Create secure enclave for master key
	buffer := memguard.NewBufferFromBytes(mk)
	ac.mk = buffer.Seal()
	buffer.Destroy()  // Clear the temporary buffer
	
	return nil
}

func (ac *ArgonCrypt) getMKUnlocked() (*memguard.LockedBuffer, error) {
	if ac.mk == nil {
		return nil, errors.New("master key not set")
	}
	return ac.mk.Open(), nil
}

func (ac *ArgonCrypt) ClearMK() {
	if ac.mk != nil {
		ac.mk.Destroy()
		ac.mk = nil
	}
}
```

**Benefits of memguard:**
- Automatic secure memory management
- Memory locked (prevents swapping to disk)
- Automatic zeroing on destruction
- Cross-platform support
- Battle-tested in security applications

**Drawbacks:**
- External dependency
- More complex API
- May be overkill for a simple CLI tool

#### Recommended Approach

For your MSK password manager, I recommend **Option B (Architecture-Based)** because:

1. ✅ No external dependencies
2. ✅ Simpler than memguard
3. ✅ Single source of truth - easier to reason about
4. ✅ Uses references, minimizing copies
5. ✅ Clear ownership model
6. ✅ Simple zeroing is sufficient (no runtime.KeepAlive needed)

#### Step 2: Update ArgonCrypt to Support Clearing Master Key

**Update: `internal/encryption/encryption.go`**

```go
package encryption

import (
	"sync"
	"github.com/amauribechtoldjr/msk/internal/domain"
)

type ArgonCrypt struct {
	mk []byte
	mu sync.RWMutex  // Protect against concurrent access
}

func NewArgonCrypt() *ArgonCrypt {
	return &ArgonCrypt{}
}

func (ac *ArgonCrypt) ConfigMK(mk []byte) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	
	// Clear old master key first
	ac.clearMKUnsafe()
	
	// Make OWNED copy - we control this memory
	ac.mk = make([]byte, len(mk))
	copy(ac.mk, mk)
	
	// Original mk in caller is separate - they clear it
}

func (ac *ArgonCrypt) clearMKUnsafe() {
	if ac.mk != nil {
		SecureZero(ac.mk)
		ac.mk = nil
	}
}

// ClearMK clears the master key from memory
func (ac *ArgonCrypt) ClearMK() {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.clearMKUnsafe()
}
```

#### Step 3: Clear Derived Keys After Encryption

**Update: `internal/encryption/encrypt.go`**

```go
func (a *ArgonCrypt) Encrypt(secret domain.Secret) (domain.EncryptedSecret, error) {
	salt, err := randomBytes(MSK_SALT_SIZE)
	if err != nil {
		return domain.EncryptedSecret{}, err
	}

	// Get master key reference (not a copy)
	mk := a.getMK()
	
	// Derive key
	key := getArgonDeriveKey(mk, salt)
	defer ZeroBytes(key) // Clear derived key after use

	block, err := aes.NewCipher(key)
	if err != nil {
		return domain.EncryptedSecret{}, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return domain.EncryptedSecret{}, err
	}

	nonce, err := randomBytes(MSK_NONCE_SIZE)
	if err != nil {
		return domain.EncryptedSecret{}, err
	}

	// Marshal to JSON
	plaintext, err := json.Marshal(secret)
	if err != nil {
		return domain.EncryptedSecret{}, err
	}
	defer ZeroBytes(plaintext) // Clear plaintext after encryption

	// Encrypt
	cipherText := gcm.Seal(nil, nonce, plaintext, nil)

	// plaintext and key will be cleared by defer

	return domain.EncryptedSecret{
		Data:  cipherText,
		Salt:  [MSK_SALT_SIZE]byte(salt),
		Nonce: [MSK_NONCE_SIZE]byte(nonce),
	}, nil
}
```

**Note:** We use `a.getMK()` to get a reference to the master key, not a copy. The derived key is a local variable that gets cleared.

#### Step 4: Clear Master Key After Operations Complete

**Update: `internal/app/service.go`**

```go
type MSKService interface {
	AddSecret(ctx context.Context, name string, password []byte) error
	GetSecret(ctx context.Context, name string) (string, error)
	DeleteSecret(ctx context.Context, name string) error
	ListAll(ctx context.Context) error
	ConfigMK(ctx context.Context, mk []byte)
	ClearMasterKey() // Add this method
}

// Implementation
func (s *Service) ClearMasterKey() {
	if crypt, ok := s.crypto.(*encryption.ArgonCrypt); ok {
		crypt.ClearMK()
	}
}
```

**Update: `internal/cli/root.go`**

```go
import (
	"github.com/amauribechtoldjr/msk/internal/app"
	"github.com/amauribechtoldjr/msk/internal/encryption"
	"github.com/spf13/cobra"
)

func NewMSKCmd(service app.MSKService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "msk",
		// ... existing config ...
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			mk, err := PromptPassword("Enter master password:")
			if err != nil {
				return err
			}
			defer encryption.ZeroBytes(mk) // Clear prompt buffer after ConfigMK takes ownership
			
			service.ConfigMK(ctx, mk)  // Service takes ownership, makes its own copy
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			// Clear master key from encryption service after command completes
			if svc, ok := service.(*app.Service); ok {
				svc.ClearMasterKey()
			}
		},
	}
	// ... rest of command setup ...
}
```

**Design Pattern:**
- CLI layer: Owns prompt buffer, clears it after passing to service
- Service layer: Takes ownership via ConfigMK, makes its own copy
- Encryption layer: Owns the master key, clears it when done
- Each layer clears what it owns - clear ownership model

#### Step 5: Clear Passwords After Use in Service Layer

**Update: `internal/app/service.go` - AddSecret**

```go
import (
	"github.com/amauribechtoldjr/msk/internal/encryption"
	// ... other imports
)

func (s *Service) AddSecret(ctx context.Context, name string, rawP []byte) error {
	// Create a copy so we can safely clear it later
	// We own this copy, so we're responsible for clearing it
	password := make([]byte, len(rawP))
	copy(password, rawP)
	
	defer encryption.ZeroBytes(password) // Clear our copy when done
	
	exists, err := s.repo.FileExists(ctx, name)
	if err != nil {
		return err
	}
	
	if exists {
		return ErrSecretExists
	}
	
	secret := domain.Secret{
		Name:      name,
		Password:  password, // Pass to encryption
		CreatedAt: time.Now().UTC(),
	}
	
	encryptionResult, err := s.crypto.Encrypt(secret)
	if err != nil {
		return err
	}
	
	// password will be cleared by defer
	
	return s.repo.SaveFile(ctx, encryptionResult, name)
}
```

**Update: `internal/app/service.go` - GetSecret**

```go
func (s *Service) GetSecret(ctx context.Context, name string) ([]byte, error) {
	// ... existing validation ...
	
	secretData, err := s.crypto.Decrypt(fileData)
	if err != nil {
		return nil, err
	}
	defer encryption.ZeroBytes(secretData.Password) // Clear after converting to string
	
	// Copy to return value - caller will own this
	password := make([]byte, len(secretData.Password))
	copy(password, secretData.Password)
	
	return password, nil
}
```

**Note:** In GetSecret, we return a copy so the caller owns the memory. The CLI layer should clear it after use (copying to clipboard).

### Testing Your Implementation

After implementing, test with Process Hacker:

1. Use unique test password: `TEST_ZERO_XYZ123`
2. Run: `msk add test -p "TEST_ZERO_XYZ123"`
3. Wait 2-3 seconds after completion
4. Search Process Hacker memory for `TEST_ZERO_XYZ123`
5. Should NOT be found if zeroing works correctly

### Pattern to Learn

**Single Ownership with Reference Passing:**
- Each component owns the memory it creates
- Pass by reference when possible to avoid copies
- Clear what you own, when you're done with it
- Clear ownership model makes code easier to reason about

**Benefits over runtime.KeepAlive approach:**
- ✅ No compiler hacks needed
- ✅ Cleaner, more maintainable code
- ✅ Clear ownership semantics
- ✅ Minimizes copies (better for security)
- ✅ Easier to understand and reason about

**Defensive Programming for Cryptographic Operations:** Assume memory can be read. Clear sensitive data immediately after use. Use clear ownership models rather than fighting the compiler.

---

## Group 2: Cryptographic Security - Argon2 Configuration

**Priority:** 🟠 **HIGH** - Security vulnerability

### Issues

1. **Weak Argon2 parameters** (time=3, memory=64KB is extremely low)
2. **Using `argon2.IDKey` instead of `argon2id`** (less resistant to GPU attacks)

### Learning Topics

- Password-based key derivation functions (PBKDF)
- Argon2 variants (argon2id vs argon2i vs argon2d)
- Balancing security vs performance
- Modern recommendations for Argon2 parameters

### Implementation

**Update: `internal/encryption/key.go`**

```go
package encryption

import "golang.org/x/crypto/argon2"

func getArgonDeriveKey(password, salt []byte) []byte {
	return argon2.IDKey(
		password,
		salt,
		2,              // time: 2-3 is acceptable
		64*1024*1024,   // memory: 64MB (up from 64KB!)
		4,              // threads
		32,             // key length
	)
}
```

**Better: Use argon2id variant**

```go
import "golang.org/x/crypto/argon2"

func getArgonDeriveKey(password, salt []byte) []byte {
	return argon2.IDKey(  // This is already argon2id, but consider making explicit
		password,
		salt,
		2,              // time: 2-3 is good
		64*1024*1024,   // memory: 64MB for interactive use
		4,              // threads
		32,             // key length: 32 bytes = 256 bits
	)
}
```

**Note:** `argon2.IDKey` in golang.org/x/crypto is actually `argon2id`. The naming is confusing but correct.

### Recommended Parameters

- **Interactive (CLI)**: time=2-3, memory=64-128MB, threads=4
- **Server**: time=2-3, memory=256-512MB, threads=4

### Pattern to Learn

**Cryptographic Primitives:** Understand the security vs performance tradeoffs. Never use weak parameters for production.

---

## Group 3: Go Error Handling & Defensive Programming

**Priority:** 🟠 **HIGH** - Code reliability

### Issues

1. **Ignored errors in `main.go`** (store creation), `add.go` (flag parsing)
2. **Error handling patterns inconsistent**
3. **Some crypto operations ignore errors** (though they rarely fail)

### Learning Topics

- Go error handling philosophy ("errors are values")
- Defensive programming - always check errors
- Error wrapping with `fmt.Errorf()` and `errors.Wrap()`
- When to handle vs propagate errors

### Implementation

**Fix: `cmd/msk/main.go`**

```go
func main() {
	store, err := file.NewStore("./vault/")
	if err != nil {
		logger.RenderError(fmt.Errorf("failed to create store: %w", err))
		os.Exit(1)
	}
	
	// ... rest of code
}
```

**Fix: `internal/cli/add.go`**

```go
value, err := cmd.Flags().GetString("password")
if err != nil {
	return fmt.Errorf("failed to parse password flag: %w", err)
}
password := []byte(value)
```

### Pattern to Learn

**"If you ignore an error, be very explicit about why"** - Prefer explicit handling over ignoring.

---

## Group 4: Input Validation & Sanitization

**Priority:** 🟡 **MEDIUM-HIGH** - Security & reliability

### Issues

1. **No input validation on secret names** (path traversal risk: `../`, `/`)
2. **No length limits on passwords or names**
3. **No character restrictions to prevent filesystem issues**

### Learning Topics

- Input validation principles
- Path traversal attacks
- Sanitization vs validation
- Filesystem security

### Implementation

**Create: `internal/domain/validation.go`**

```go
package domain

import (
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalidSecretName = errors.New("invalid secret name")
	ErrNameTooLong       = errors.New("secret name too long")
)

const (
	MaxSecretNameLength = 255
)

func ValidateSecretName(name string) error {
	if name == "" {
		return ErrInvalidSecretName
	}
	
	if utf8.RuneCountInString(name) > MaxSecretNameLength {
		return ErrNameTooLong
	}
	
	// Prevent path traversal and special characters
	if strings.Contains(name, "..") ||
		strings.Contains(name, "/") ||
		strings.Contains(name, "\\") ||
		strings.Contains(name, "\x00") {
		return ErrInvalidSecretName
	}
	
	// Prevent reserved names on Windows
	reserved := []string{"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	upperName := strings.ToUpper(name)
	for _, r := range reserved {
		if upperName == r {
			return ErrInvalidSecretName
		}
	}
	
	return nil
}
```

**Use in service layer:**

```go
func (s *Service) AddSecret(ctx context.Context, name string, rawP []byte) error {
	if err := domain.ValidateSecretName(name); err != nil {
		return err
	}
	// ... rest of code
}
```

### Pattern to Learn

**"Validate inputs at boundaries"** - Validate at the service/CLI boundary.

---

## Group 5: Context Usage & Cancellation

**Priority:** 🟡 **MEDIUM** - Go best practices

### Issues

1. **Context passed but not fully utilized**
2. **No cancellation support for long-running operations** (Argon2)
3. **Context checked but not propagated to all layers**

### Learning Topics

- Context package in Go
- Cancellation and timeouts
- Context propagation patterns
- When to use context

### Implementation

Note: Argon2 operations can't be easily cancelled, but you can add timeouts:

```go
func (s *Service) GetSecret(ctx context.Context, name string) (string, error) {
	// Add timeout for long operations
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	// ... existing code
}
```

### Pattern to Learn

Context is for cancellation, timeouts, and request-scoped values - use it appropriately.

---

## Group 6: Service Layer Design & Separation of Concerns

**Priority:** 🟡 **MEDIUM** - Architecture & maintainability

### Issues

1. **Service layer leaks implementation** (returns `[]byte` instead of domain types)
2. **Mixed concerns** (service layer shouldn't know about clipboard)
3. **Service interface returns concrete type** (`*Service` instead of interface)

### Learning Topics

- Clean Architecture / Layered Architecture
- Interface design (Interface Segregation Principle)
- Dependency Inversion
- Separation of Concerns

### Implementation

**Current (mixed concerns):**
```go
func (s *Service) GetSecret(...) (string, error) {
	// ... decrypt ...
	return password, nil  // Service returns string
}
```

**Better (separation):**
```go
func (s *Service) GetSecret(...) (string, error) {
	// Service just returns the password
	// CLI layer handles clipboard copying
	return password, nil
}
```

The CLI layer (not service) should handle clipboard operations - this is already correct!

### Pattern to Learn

**Dependency Rule:** Inner layers shouldn't depend on outer layers.

---

## Group 7: Storage & File System Security

**Priority:** 🟡 **MEDIUM** - Security & reliability

### Issues

1. **No file locking** (concurrent access risk)
2. **Atomic writes not guaranteed on all platforms** (Windows rename behavior)
3. **No secure file deletion** (deleted files recoverable)
4. **File permissions good** (0o600/0o700) but could verify parent dirs

### Learning Topics

- File locking mechanisms
- Atomic file operations
- Secure deletion
- Cross-platform file operations

### Implementation

**File locking example:**

```go
import (
	"os"
	"syscall"
)

func (s *Store) SaveFileWithLock(ctx context.Context, encryption domain.EncryptedSecret, name string) error {
	path := s.secretPath(name)
	
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600)
	if err != nil {
		if os.IsExist(err) {
			return errors.New("file already exists")
		}
		return err
	}
	defer file.Close()
	
	// Lock file (Unix/Linux)
	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
	if err != nil {
		return err
	}
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	
	// Write data
	// ...
}
```

**Note:** File locking is platform-specific. Consider using a library like `github.com/gofrs/flock` for cross-platform support.

### Pattern to Learn

File system operations are platform-dependent - test on all target platforms.

---

## Group 8: Dependency Management & Module Configuration

**Priority:** 🟢 **LOW** - Code quality

### Issues

1. **Dependencies marked as `// indirect` but should be direct**
2. **Go version in `go.mod` seems incorrect** (1.25.5 doesn't exist)

### Learning Topics

- Go modules
- Direct vs indirect dependencies
- `go mod tidy`
- Version management

### Implementation

```bash
# Fix dependencies
go mod tidy

# Fix Go version in go.mod (should be 1.21 or 1.22, not 1.25.5)
# Edit go.mod manually
```

### Pattern to Learn

Keep `go.mod` clean and accurate - run `go mod tidy` regularly.

---

## Group 9: Cobra CLI Best Practices

**Priority:** 🟢 **LOW** - UX & maintainability

### Issues

1. **Unused flags** (`--toggle`, `--master` flag defined but not read)
2. **Missing command documentation** (empty `Long` descriptions)
3. **Manual argument validation** instead of Cobra's built-in `Args`
4. **Missing examples in help text**

### Learning Topics

- Cobra framework best practices
- CLI UX design
- Command documentation
- Argument validation patterns

### Implementation

**Use Cobra Args validation:**

```go
getCmd := &cobra.Command{
	Use:   "get <name>",
	Short: "Retrieve a password from the vault",
	Long: `Retrieve a password by name and copy it to the clipboard.
	
The password will be available for pasting (Ctrl+V) until you copy something else.
	
Examples:
  msk get github
  msk get my-secret-password`,
	Args: cobra.ExactArgs(1), // Built-in validation!
	RunE: func(cmd *cobra.Command, args []string) error {
		// args[0] is guaranteed to exist
		name := args[0]
		// ... rest of code
	},
}
```

### Pattern to Learn

Use framework features - don't reinvent validation when Cobra provides it.

---

## Group 10: Configuration & Flexibility

**Priority:** 🟡 **MEDIUM** - Usability

### Issues

1. **Hardcoded vault path** (`./vault/`)
2. **No configuration system**
3. **Argon2 parameters hardcoded**
4. **TODO comment acknowledges missing config**

### Learning Topics

- Configuration management patterns
- XDG directory conventions
- Environment variables
- Config file formats (TOML, YAML, JSON)

### Implementation

**Support XDG directories:**

```go
package config

import (
	"os"
	"path/filepath"
)

func GetVaultPath() (string, error) {
	// Check environment variable first
	if vault := os.Getenv("MSK_VAULT_PATH"); vault != "" {
		return vault, nil
	}
	
	// Use XDG data directory
	if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
		return filepath.Join(xdgData, "msk", "vault"), nil
	}
	
	// Fallback to home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	// Windows: AppData\Local\msk\vault
	// Unix: ~/.local/share/msk/vault
	if os.PathSeparator == '\\' {
		return filepath.Join(home, "AppData", "Local", "msk", "vault"), nil
	}
	return filepath.Join(home, ".local", "share", "msk", "vault"), nil
}
```

### Pattern to Learn

**"Convention over Configuration"** - Use standard locations (XDG) but allow overrides.

---

## Group 11: Missing Features & Incomplete Implementation

**Priority:** 🟡 **LOW-MEDIUM** - Feature completeness

### Issues

1. **`ListAll()` not implemented** (returns `nil`)
2. **Missing unit tests** (entire codebase)
3. **Error message bugs** (get command says "Password added successfully")

### Learning Topics

- Test-driven development
- Unit testing in Go
- Integration testing
- Mocking interfaces

### Implementation

**Implement ListAll:**

```go
func (s *Service) ListAll(ctx context.Context) ([]string, error) {
	// Read vault directory
	// Return list of secret names (without decrypting)
	// Implementation depends on your storage backend
}
```

**Add unit tests:**

```go
// internal/app/service_test.go
package app

import (
	"context"
	"testing"
)

func TestService_AddSecret(t *testing.T) {
	// Mock repository and encryption
	// Test adding a secret
	// Verify it was saved correctly
}
```

### Pattern to Learn

Testing - start with interfaces, make things testable, write tests.

---

## Recommended Learning Order

### Week 1: Critical Security Foundations
1. **Group 1: Memory Safety** (most critical)
2. **Group 2: Argon2 Configuration** (security vulnerability)

### Week 2: Code Quality & Reliability
3. **Group 3: Error Handling**
4. **Group 4: Input Validation**

### Week 3: Architecture & Design
5. **Group 5: Context Usage**
6. **Group 6: Service Layer Design**

### Week 4: Infrastructure & Polish
7. **Group 7: Storage Security**
8. **Group 10: Configuration System**
9. **Group 11: Missing Features** (tests, ListAll)

### Ongoing (Lower Priority)
- **Group 8: Dependency Management** (quick fix)
- **Group 9: Cobra Improvements** (when improving UX)

---

## Key Security Principles to Remember

1. **Defense in Depth**: Multiple layers of security
2. **Fail Secure**: If something fails, fail in a secure way
3. **Minimize Attack Surface**: Don't store/transmit more than necessary
4. **Principle of Least Privilege**: Give minimum access needed
5. **Secure by Default**: Secure defaults, require opt-in for weaker options

---

## Key Go Design Patterns to Learn

1. **Error Handling**: Errors are values, handle them explicitly
2. **Interfaces**: Program to interfaces, not implementations
3. **Context**: Use for cancellation, timeouts, request-scoped values
4. **Composition**: Prefer composition over inheritance
5. **Package Design**: Clear boundaries, minimal exports

---

## Testing Your Improvements

### Memory Safety Testing

Use Process Hacker to verify memory cleanup:

1. Run your MSK program
2. Perform operations (add/get secrets)
3. Use Process Hacker to search memory for test passwords
4. Verify sensitive data is NOT found after operations complete

### Security Testing Checklist

- [ ] Master key not found in memory after command completion
- [ ] Passwords not found in memory after encryption
- [ ] Derived keys cleared after use
- [ ] Input validation prevents path traversal
- [ ] File permissions are correct (0o600 for files, 0o700 for directories)
- [ ] Argon2 parameters are strong enough

---

## Resources

- [Go Security Best Practices](https://go.dev/security/best-practices)
- [OWASP Secure Coding Practices](https://owasp.org/www-project-secure-coding-practices-quick-reference-guide/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

---

*Last Updated: Based on code review conducted on the MSK password manager codebase.*
