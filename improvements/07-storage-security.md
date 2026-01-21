# Group 7: Storage and File System Security

**Priority:** 🟡 **MEDIUM** - Security and reliability (Phase 3)

## Current State

The storage layer has some good practices:
- Uses 0o600 permissions for files (owner read/write only)
- Uses 0o700 permissions for directories
- Implements atomic writes with temp file + rename

However, there are gaps to address.

## Issues

1. **No file locking** - Concurrent access could corrupt data
2. **Atomic writes not guaranteed on Windows** - Rename behavior differs
3. **No secure file deletion** - Deleted files are recoverable
4. **Directory permissions not verified** - Parent dirs could be insecure

## Learning Topics

- File locking mechanisms (flock, LockFileEx)
- Atomic file operations across platforms
- Secure deletion concepts
- Cross-platform file operations

## Implementation

### Step 1: Add File Locking

File locking prevents concurrent access from corrupting vault files.

**Option A: Use a library (Recommended)**

```bash
go get github.com/gofrs/flock
```

**Update: `internal/storage/file/store.go`**

```go
package file

import (
    "context"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/gofrs/flock"
    "github.com/amauribechtoldjr/msk/internal/domain"
)

var (
    ErrNotFound    = errors.New("secret not found")
    ErrLockTimeout = errors.New("could not acquire file lock")
)

const (
    lockTimeout = 5 * time.Second
)

type Store struct {
    dir string
}

func NewStore(dir string) (*Store, error) {
    if err := os.MkdirAll(dir, 0o700); err != nil {
        return nil, fmt.Errorf("failed to create vault directory: %w", err)
    }
    
    // Verify directory permissions
    info, err := os.Stat(dir)
    if err != nil {
        return nil, err
    }
    
    // On Unix, check permissions are restrictive
    // This is a best-effort check - Windows permissions work differently
    if info.Mode().Perm()&0o077 != 0 {
        // Directory is accessible by group/others
        // Try to fix it
        if err := os.Chmod(dir, 0o700); err != nil {
            return nil, fmt.Errorf("vault directory has insecure permissions: %w", err)
        }
    }
    
    return &Store{dir: dir}, nil
}

func (s *Store) SaveFile(ctx context.Context, encryption domain.EncryptedSecret, name string) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    path := s.secretPath(name)
    lockPath := path + ".lock"
    
    // Create file lock
    fileLock := flock.New(lockPath)
    
    // Try to acquire lock with timeout
    ctx, cancel := context.WithTimeout(ctx, lockTimeout)
    defer cancel()
    
    locked, err := fileLock.TryLockContext(ctx, 100*time.Millisecond)
    if err != nil {
        return fmt.Errorf("failed to acquire lock: %w", err)
    }
    if !locked {
        return ErrLockTimeout
    }
    defer fileLock.Unlock()
    defer os.Remove(lockPath) // Clean up lock file
    
    // Write to temp file first (atomic write)
    tmpPath := path + ".tmp"
    
    tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
    if err != nil {
        return fmt.Errorf("failed to create temp file: %w", err)
    }
    
    // Ensure cleanup on error
    success := false
    defer func() {
        tmpFile.Close()
        if !success {
            os.Remove(tmpPath)
        }
    }()
    
    // Write header and data
    if _, err := tmpFile.Write([]byte("MSK")); err != nil {
        return err
    }
    if _, err := tmpFile.Write([]byte{1}); err != nil {
        return err
    }
    if _, err := tmpFile.Write(encryption.Salt[:]); err != nil {
        return err
    }
    if _, err := tmpFile.Write(encryption.Nonce[:]); err != nil {
        return err
    }
    if _, err := tmpFile.Write(encryption.Data); err != nil {
        return err
    }
    
    // Sync to disk
    if err := tmpFile.Sync(); err != nil {
        return err
    }
    if err := tmpFile.Close(); err != nil {
        return err
    }
    
    // Atomic rename
    if err := os.Rename(tmpPath, path); err != nil {
        return fmt.Errorf("failed to finalize file: %w", err)
    }
    
    // Sync directory (Unix only, Windows ignores this)
    s.syncDir()
    
    success = true
    return nil
}

func (s *Store) GetFile(ctx context.Context, name string) ([]byte, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    path := s.secretPath(name)
    lockPath := path + ".lock"
    
    // Shared lock for reading
    fileLock := flock.New(lockPath)
    
    ctx, cancel := context.WithTimeout(ctx, lockTimeout)
    defer cancel()
    
    locked, err := fileLock.TryRLockContext(ctx, 100*time.Millisecond)
    if err != nil {
        return nil, fmt.Errorf("failed to acquire read lock: %w", err)
    }
    if !locked {
        return nil, ErrLockTimeout
    }
    defer fileLock.Unlock()
    
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, ErrNotFound
        }
        return nil, err
    }
    
    return data, nil
}

func (s *Store) DeleteFile(ctx context.Context, name string) (bool, error) {
    select {
    case <-ctx.Done():
        return false, ctx.Err()
    default:
    }
    
    path := s.secretPath(name)
    lockPath := path + ".lock"
    
    // Exclusive lock for deletion
    fileLock := flock.New(lockPath)
    
    ctx, cancel := context.WithTimeout(ctx, lockTimeout)
    defer cancel()
    
    locked, err := fileLock.TryLockContext(ctx, 100*time.Millisecond)
    if err != nil {
        return false, fmt.Errorf("failed to acquire lock: %w", err)
    }
    if !locked {
        return false, ErrLockTimeout
    }
    defer fileLock.Unlock()
    defer os.Remove(lockPath)
    
    err = os.Remove(path)
    if err != nil {
        if os.IsNotExist(err) {
            return false, ErrNotFound
        }
        return false, err
    }
    
    return true, nil
}

func (s *Store) FileExists(ctx context.Context, name string) (bool, error) {
    select {
    case <-ctx.Done():
        return false, ctx.Err()
    default:
    }
    
    _, err := os.Stat(s.secretPath(name))
    if err == nil {
        return true, nil
    }
    if os.IsNotExist(err) {
        return false, nil
    }
    return false, err
}

func (s *Store) ListFiles(ctx context.Context) ([]string, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    entries, err := os.ReadDir(s.dir)
    if err != nil {
        return nil, fmt.Errorf("failed to read vault: %w", err)
    }
    
    var names []string
    for _, entry := range entries {
        if entry.IsDir() {
            continue
        }
        name := entry.Name()
        if strings.HasSuffix(name, ".msk") {
            names = append(names, strings.TrimSuffix(name, ".msk"))
        }
    }
    
    return names, nil
}

func (s *Store) secretPath(name string) string {
    return filepath.Join(s.dir, name+".msk")
}

func (s *Store) syncDir() {
    dir, err := os.Open(s.dir)
    if err != nil {
        return
    }
    defer dir.Close()
    _ = dir.Sync() // Best effort
}
```

### Option B: Manual File Locking (No Dependencies)

If you prefer not to add a dependency, here's a simpler approach using lock files:

```go
// Simple lock file approach
func (s *Store) acquireLock(name string) (func(), error) {
    lockPath := s.secretPath(name) + ".lock"
    
    // Try to create lock file exclusively
    for i := 0; i < 50; i++ { // 5 second timeout
        f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL, 0o600)
        if err == nil {
            f.Close()
            return func() { os.Remove(lockPath) }, nil
        }
        time.Sleep(100 * time.Millisecond)
    }
    
    return nil, ErrLockTimeout
}
```

## Secure Deletion Note

True secure deletion (overwriting data multiple times) is complex because:
- SSDs have wear leveling that makes overwriting unreliable
- File systems may keep copies in journals
- Most modern systems use encryption anyway

For MSK, since the vault files are encrypted, standard deletion is acceptable. The master key protection is more important.

## Pattern to Learn

**File System Security:**
- Use file locking to prevent concurrent access corruption
- Atomic writes: write to temp file, then rename
- Set restrictive permissions (0o600 for files, 0o700 for directories)
- Consider platform differences (Windows vs Unix)

**Cross-Platform Considerations:**
- File locking works differently on Windows vs Unix
- Use libraries like `github.com/gofrs/flock` for portability
- Test on all target platforms
