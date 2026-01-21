# Group 2: Cryptographic Security - Argon2 Configuration

**Priority:** 🟠 **HIGH** - Security vulnerability (Phase 3)

## Current Problem

**File:** `internal/encryption/key.go`

The current Argon2 parameters are dangerously weak:

```go
// CURRENT (WEAK) - DO NOT USE IN PRODUCTION
return argon2.IDKey(
    password,
    salt,
    3,      // time iterations
    64,     // memory: 64 KB - EXTREMELY LOW!
    4,      // threads
    32,     // key length
)
```

**Why 64KB is dangerous:**
- Modern password crackers can test billions of passwords per second with such low memory
- GPU-based attacks are highly effective against low-memory Argon2
- Even a modest gaming GPU could crack weak passwords in minutes

## Issues

1. **Weak Argon2 parameters** - memory=64KB is 1000x lower than recommended
2. **Confusion about argon2.IDKey** - This IS argon2id (the recommended variant)

## Learning Topics

- Password-based key derivation functions (PBKDF)
- Argon2 variants: argon2id (recommended), argon2i, argon2d
- Memory-hard functions and why they matter
- Balancing security vs performance for CLI applications

## Implementation

**Update: `internal/encryption/key.go`**

```go
package encryption

import "golang.org/x/crypto/argon2"

// Argon2 parameters for interactive CLI use
// These provide good security while keeping the delay under 1 second
const (
    argonTime    = 2           // Number of iterations
    argonMemory  = 64 * 1024   // 64 MB in KB (was 64 KB!)
    argonThreads = 4           // Parallel threads
    argonKeyLen  = 32          // 256-bit key
)

// getArgonDeriveKey derives an encryption key from a password and salt.
// Uses Argon2id which is resistant to both side-channel and GPU attacks.
func getArgonDeriveKey(password, salt []byte) []byte {
    return argon2.IDKey(
        password,
        salt,
        argonTime,
        argonMemory,
        argonThreads,
        argonKeyLen,
    )
}
```

## Parameter Recommendations

| Use Case | Time | Memory | Threads | Notes |
|----------|------|--------|---------|-------|
| Interactive CLI | 2-3 | 64 MB | 4 | Good for user-facing apps |
| Server/API | 2-3 | 256 MB | 4 | Higher security for servers |
| High security | 3-4 | 1 GB | 4 | When delay is acceptable |
| Minimum acceptable | 2 | 19 MB | 1 | OWASP minimum (not recommended) |

## Why argon2.IDKey is Correct

The Go function `argon2.IDKey` implements **Argon2id**, which is the recommended variant:

- **Argon2d**: Fastest, but vulnerable to side-channel attacks
- **Argon2i**: Side-channel resistant, but weaker against GPU attacks
- **Argon2id**: Hybrid - combines best of both (recommended by OWASP)

The naming `IDKey` stands for "**ID** variant **Key** derivation" - it IS argon2id.

## Testing the Change

After updating, you can verify the delay is acceptable:

```go
// Add this temporarily to test
import "time"

start := time.Now()
key := getArgonDeriveKey([]byte("test"), []byte("salt1234salt1234"))
fmt.Printf("Key derivation took: %v\n", time.Since(start))
// Should be 200ms - 1s depending on your hardware
```

## Migration Consideration

**Warning:** Changing Argon2 parameters will make existing vault files unreadable!

Options:
1. **Clean break**: Delete old vault files, start fresh (simplest for learning)
2. **Versioned migration**: Store parameters in file header, support both old and new
3. **Re-encryption tool**: Decrypt with old params, re-encrypt with new

For learning purposes, option 1 is recommended.

## Pattern to Learn

**Cryptographic Primitives:**
- Always use recommended parameters from security organizations (OWASP, NIST)
- Memory-hard functions protect against GPU/ASIC attacks
- Test performance on target hardware before deploying
- Never use weak parameters in production - even for "learning projects" that might be used later
