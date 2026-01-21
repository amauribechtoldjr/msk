# Group 1: Memory Safety and Sensitive Data Handling

**Priority:** 🔴 **CRITICAL** - But do this LAST (Phase 5)

**Why last?** This is the most complex improvement, requiring understanding of Go memory model, platform-specific system calls, and compiler optimizations. Complete all other phases first to have a solid foundation.

## Issues

1. **Master key never cleared from memory** (`internal/encryption/encryption.go:30`)
2. **Passwords stored in memory but not cleared after use**
3. **Derived encryption keys not cleared after encryption** (only after decryption)
4. **Password output security** (Already partially fixed - now using clipboard)

## Learning Topics

- Secure memory management in Go
- Implementing memguard-like functionality manually (for learning)
- Preventing sensitive data exposure in memory dumps
- Memory locking to prevent swapping to disk
- Using `runtime.KeepAlive()` to prevent compiler optimizations
- Zeroing byte slices securely with compiler optimization prevention

## Approach: Manual Implementation of Memguard-Like Functionality

This implementation teaches you how secure memory libraries like memguard work by building similar functionality from scratch. This is a learning exercise that will help you understand:

1. How memory locking works (mlock/VirtualAlloc)
2. How to prevent compiler optimizations
3. How to ensure memory is cleared
4. The challenges in secure memory management

## Implementation: Secure Memory Package

### Step 1: Create Secure Buffer Package

**Create: `internal/secure/secure.go`**

This package implements secure memory management similar to memguard:

```go
package secure

import (
	"runtime"
	"sync"
	"unsafe"
)

// SecureBuffer provides secure memory management for sensitive data.
// It ensures memory is locked (prevented from swapping to disk) and
// cleared on destruction. This is a learning implementation of what
// libraries like memguard do.
type SecureBuffer struct {
	data   []byte
	locked bool
	mu     sync.RWMutex
}

// NewSecureBuffer creates a new secure buffer from bytes.
// The buffer will be locked in memory and cleared on destruction.
func NewSecureBuffer(data []byte) (*SecureBuffer, error) {
	sb := &SecureBuffer{
		data: make([]byte, len(data)),
	}
	copy(sb.data, data)
	
	// Lock memory to prevent swapping (platform-specific)
	if err := sb.lockMemory(); err != nil {
		// Clear on error - don't leave sensitive data around
		secureZero(sb.data)
		return nil, err
	}
	
	sb.locked = true
	
	// Register finalizer to ensure clearing if Destroy() is not called
	runtime.SetFinalizer(sb, (*SecureBuffer).Destroy)
	
	return sb, nil
}

// Bytes returns a reference to the underlying data.
// WARNING: Caller must not modify the returned slice.
// Use Copy() if you need a modifiable copy.
func (sb *SecureBuffer) Bytes() []byte {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	if sb.data == nil {
		return nil
	}
	return sb.data
}

// Copy creates a copy of the buffer's data.
// The caller owns the returned slice and should clear it with ZeroBytes().
func (sb *SecureBuffer) Copy() []byte {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	if sb.data == nil {
		return nil
	}
	result := make([]byte, len(sb.data))
	copy(result, sb.data)
	return result
}

// Len returns the length of the buffer.
func (sb *SecureBuffer) Len() int {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return len(sb.data)
}

// Destroy clears and unlocks the memory.
// Should be called explicitly when done with the buffer.
// Safe to call multiple times (idempotent).
func (sb *SecureBuffer) Destroy() {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	
	if sb.data == nil {
		return // Already destroyed
	}
	
	// Clear the memory securely
	secureZero(sb.data)
	
	// Unlock memory (platform-specific)
	if sb.locked {
		sb.unlockMemory()
		sb.locked = false
	}
	
	// Clear the slice reference
	sb.data = nil
	
	// Remove finalizer (no longer needed)
	runtime.SetFinalizer(sb, nil)
}

// secureZero clears memory securely, preventing compiler optimizations.
// This is critical - simple loops can be optimized away by the compiler.
func secureZero(data []byte) {
	if len(data) == 0 {
		return
	}
	
	// Use unsafe to get pointer to first byte
	ptr := unsafe.Pointer(&data[0])
	
	// Clear using memclr (similar to memset)
	memclrNoHeapPointers(ptr, uintptr(len(data)))
	
	// CRITICAL: Prevent compiler optimization
	// Without this, the compiler might remove the clearing loop
	runtime.KeepAlive(data)
}

// memclrNoHeapPointers clears memory.
// This is a simplified version - real implementations use assembly for speed.
// For learning, we use a loop with KeepAlive to prevent optimization.
func memclrNoHeapPointers(ptr unsafe.Pointer, n uintptr) {
	if n == 0 {
		return
	}
	
	// Convert pointer to byte slice
	bytes := (*[1 << 30]byte)(ptr)[:n:n]
	
	// Clear each byte
	for i := range bytes {
		bytes[i] = 0
	}
	
	// CRITICAL: Keep pointer alive to prevent optimization
	runtime.KeepAlive(ptr)
}

// ZeroBytes is a helper function to zero regular byte slices.
// Use this for temporary buffers that aren't in SecureBuffers.
// This uses the same secure zeroing approach.
func ZeroBytes(b []byte) {
	if len(b) == 0 {
		return
	}
	
	ptr := unsafe.Pointer(&b[0])
	memclrNoHeapPointers(ptr, uintptr(len(b)))
	runtime.KeepAlive(b)
}
```

### Step 2: Platform-Specific Memory Locking

**Create: `internal/secure/lock_unix.go`** (for Unix/Linux/macOS)

```go
//go:build linux || darwin || freebsd

package secure

import (
	"syscall"
	"unsafe"
)

// lockMemory locks the buffer's memory pages to prevent swapping to disk.
// Uses mlock() system call on Unix systems.
func (sb *SecureBuffer) lockMemory() error {
	if len(sb.data) == 0 {
		return nil
	}
	
	ptr := unsafe.Pointer(&sb.data[0])
	
	// Call mlock to lock memory pages
	// This prevents the OS from swapping this memory to disk
	_, _, errno := syscall.Syscall(
		syscall.SYS_MLOCK,
		uintptr(ptr),
		uintptr(len(sb.data)),
		0,
	)
	
	if errno != 0 {
		// mlock can fail if:
		// - Process doesn't have privilege (RLIMIT_MEMLOCK)
		// - System doesn't support it
		// For learning, we'll continue but log the issue
		// In production, you might want to handle this differently
		return errno
	}
	
	return nil
}

// unlockMemory unlocks the buffer's memory pages.
// Uses munlock() system call on Unix systems.
func (sb *SecureBuffer) unlockMemory() error {
	if len(sb.data) == 0 {
		return nil
	}
	
	ptr := unsafe.Pointer(&sb.data[0])
	
	// Call munlock to unlock memory pages
	_, _, errno := syscall.Syscall(
		syscall.SYS_MUNLOCK,
		uintptr(ptr),
		uintptr(len(sb.data)),
		0,
	)
	
	if errno != 0 {
		return errno
	}
	
	return nil
}
```

**Create: `internal/secure/lock_windows.go`** (for Windows)

```go
//go:build windows

package secure

import (
	"syscall"
	"unsafe"
)

var (
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	virtualLockProc   = kernel32.NewProc("VirtualLock")
	virtualUnlockProc = kernel32.NewProc("VirtualUnlock")
)

// lockMemory locks the buffer's memory pages on Windows.
// Uses VirtualLock() API call.
func (sb *SecureBuffer) lockMemory() error {
	if len(sb.data) == 0 {
		return nil
	}
	
	ptr := unsafe.Pointer(&sb.data[0])
	
	// Call VirtualLock to lock memory pages
	// This prevents the OS from swapping this memory to disk
	ret, _, errno := virtualLockProc.Call(
		uintptr(ptr),
		uintptr(len(sb.data)),
	)
	
	if ret == 0 {
		// VirtualLock can fail if:
		// - Process doesn't have privilege
		// - Memory limit exceeded
		// For learning, we'll continue but the operation failed
		return errno
	}
	
	return nil
}

// unlockMemory unlocks the buffer's memory pages on Windows.
// Uses VirtualUnlock() API call.
func (sb *SecureBuffer) unlockMemory() error {
	if len(sb.data) == 0 {
		return nil
	}
	
	ptr := unsafe.Pointer(&sb.data[0])
	
	// Call VirtualUnlock to unlock memory pages
	ret, _, errno := virtualUnlockProc.Call(
		uintptr(ptr),
		uintptr(len(sb.data)),
	)
	
	if ret == 0 {
		return errno
	}
	
	return nil
}
```

### Step 3: Update Encryption Layer

**Update: `internal/encryption/encryption.go`**

```go
package encryption

import (
	"sync"
	
	"github.com/amauribechtoldjr/msk/internal/domain"
	"github.com/amauribechtoldjr/msk/internal/secure"
)

const (
	MSK_MAGIC_VALUE  = "MSK"
	MSK_FILE_VERSION = byte(1)

	MSK_MAGIC_SIZE   = 3
	MSK_VERSION_SIZE = 1
	MSK_SALT_SIZE    = 16
	MSK_NONCE_SIZE   = 12
	MSK_HEADER_SIZE  = MSK_MAGIC_SIZE + MSK_VERSION_SIZE + MSK_SALT_SIZE + MSK_NONCE_SIZE
)

type Encryption interface {
	Encrypt(secret domain.Secret) (domain.EncryptedSecret, error)
	Decrypt(cipherData []byte) (domain.Secret, error)
	ConfigMK(mk *secure.SecureBuffer)
	ClearMK()
}

type ArgonCrypt struct {
	mk *secure.SecureBuffer
	mu sync.RWMutex
}

func NewArgonCrypt() *ArgonCrypt {
	return &ArgonCrypt{}
}

func (ac *ArgonCrypt) ConfigMK(mk *secure.SecureBuffer) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	
	// Clear old master key if exists
	if ac.mk != nil {
		ac.mk.Destroy()
	}
	
	ac.mk = mk
}

func (ac *ArgonCrypt) ClearMK() {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	
	if ac.mk != nil {
		ac.mk.Destroy()
		ac.mk = nil
	}
}

// getMK returns a reference to the master key bytes (internal use only)
func (ac *ArgonCrypt) getMK() []byte {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	if ac.mk == nil {
		return nil
	}
	return ac.mk.Bytes()
}
```

**Update: `internal/encryption/encrypt.go`** - Add secure clearing

```go
func (a *ArgonCrypt) Encrypt(secret domain.Secret) (domain.EncryptedSecret, error) {
	salt, err := randomBytes(MSK_SALT_SIZE)
	if err != nil {
		return domain.EncryptedSecret{}, err
	}

	mk := a.getMK()
	if mk == nil {
		return domain.EncryptedSecret{}, errors.New("master key not set")
	}
	
	// Derive key from master key and salt
	key := getArgonDeriveKey(mk, salt)
	defer secure.ZeroBytes(key) // CRITICAL: Clear derived key after use

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

	plaintext, err := json.Marshal(secret)
	if err != nil {
		return domain.EncryptedSecret{}, err
	}
	defer secure.ZeroBytes(plaintext) // CRITICAL: Clear plaintext after encryption

	cipherText := gcm.Seal(nil, nonce, plaintext, nil)

	return domain.EncryptedSecret{
		Data:  cipherText,
		Salt:  [MSK_SALT_SIZE]byte(salt),
		Nonce: [MSK_NONCE_SIZE]byte(nonce),
	}, nil
}
```

### Step 4: Update Service Layer

**Update: `internal/app/service.go`**

```go
func (s *Service) ConfigMK(ctx context.Context, mk *secure.SecureBuffer) {
	s.crypto.ConfigMK(mk)
}

func (s *Service) ClearMasterKey() {
	s.crypto.ClearMK()
}
```

### Step 5: Update CLI Layer

**Update: `internal/cli/root.go`**

```go
func NewMSKCmd(service app.MSKService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "msk",
		Short: "MSK Password Manager",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			
			passwordBytes, err := PromptPassword("Enter master password:")
			if err != nil {
				return err
			}
			defer secure.ZeroBytes(passwordBytes) // Clear prompt buffer
			
			mkBuffer, err := secure.NewSecureBuffer(passwordBytes)
			if err != nil {
				return err
			}
			
			service.ConfigMK(ctx, mkBuffer)
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			// Clear master key when command finishes
			service.ClearMasterKey()
		},
	}
	// ... add subcommands
	return cmd
}
```

## Why This Approach Works

### The Compiler Optimization Problem

The Go compiler can optimize away memory writes that appear "dead" (not used after the write). This is called "dead store elimination":

```go
// Example of problematic code:
func bad() {
	password := []byte("secret")
	// ... use password ...
	for i := range password {
		password[i] = 0  // Compiler might remove this!
	}
	// password never used again - compiler sees this as dead store
}
```

### Our Solution

1. **Use `runtime.KeepAlive()`**: Prevents the compiler from optimizing away the memory writes
2. **Use `unsafe.Pointer`**: Makes it harder for the compiler to reason about the memory
3. **Lock memory**: Prevents swapping to disk (mlock/VirtualLock)
4. **Clear in finalizer**: Ensures memory is cleared even if Destroy() isn't called
5. **Clear ownership model**: Architecture ensures we know when to clear

## Testing Your Implementation

After implementing, test with Process Hacker (Windows) or similar memory inspection tools:

1. Use unique test password: `TEST_ZERO_XYZ123`
2. Run: `msk add test -p "TEST_ZERO_XYZ123"`
3. Wait 2-3 seconds after completion
4. Search process memory for `TEST_ZERO_XYZ123`
5. Should NOT be found if zeroing works correctly

### Test Checklist

- [ ] Master key not found in memory after command completion
- [ ] Passwords not found in memory after encryption
- [ ] Derived keys cleared after use
- [ ] Memory locking works (check with system tools)
- [ ] Finalizers called correctly (if Destroy() not called)

## Pattern to Learn

**Secure Memory Management:**
- Lock memory to prevent swapping (mlock/VirtualLock)
- Clear memory securely with compiler optimization prevention
- Use clear ownership models
- Use finalizers as backup (but prefer explicit cleanup)
- Clear immediately after use - don't wait

---

**Defensive Programming for Cryptographic Operations:** Assume memory can be read. Clear sensitive data immediately after use. Use clear ownership models. Prevent compiler optimizations from defeating your security measures.
