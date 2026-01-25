# MSK Improvements Execution Plan

This plan organizes all improvements from the `improvements/` folder into a logical execution order. The goal is to ensure the project has a solid architecture and runs correctly before tackling complex memory safety topics.

**Key Principle:** Memory safety (Phase 5) is intentionally last because it requires understanding Go's memory model, platform-specific system calls, and compiler optimizations.

---

## Progress Summary

| Phase | Status | Notes |
|-------|--------|-------|
| 1.1 Error Handling | ✅ DONE | main.go and add.go now handle errors properly |
| 1.2 Input Validation | ❌ TODO | No validation.go, names not validated |
| 1.3 Missing Features | ❌ TODO | ListAll stub exists but not implemented |
| 2.1 Clean Architecture | ⚠️ PARTIAL | No cleanup hooks yet |
| 2.2 Service Layer Design | ✅ DONE | Clean separation achieved |
| 2.3 Context Usage | ✅ DONE | Context checked in storage layer |
| 3.1 Argon2 Configuration | ✅ DONE | Memory now 64MB (was 64KB) |
| 3.2 Storage Security | ✅ DONE | Atomic writes with temp files |
| 4.1 Configuration System | ❌ TODO | Vault path still hardcoded |
| 4.2 Dependency Management | ⚠️ PARTIAL | Go version 1.25.5 in go.mod (check if correct) |
| 4.3 Cobra CLI Improvements | ❌ TODO | Unused flags still present |
| 5.1 Memory Safety | ⚠️ PARTIAL | cleanupByte in decrypt.go, but incomplete |

---

## Execution Phases

### Phase 1: Foundation and Reliability

**Goal:** Ensure the project runs correctly and handles errors properly.

#### 1.1 Error Handling (03-error-handling.md) ✅ COMPLETED
- ~~cmd/msk/main.go - Handle store creation error~~ ✅ Done
- ~~internal/cli/add.go - Handle flag parsing error~~ ✅ Done
- **What was done:** Both files now properly check and wrap errors

#### 1.2 Input Validation (04-input-validation.md) ❌ TODO
- **Files to create:**
  - internal/domain/validation.go - Validation functions
- **Files to modify:**
  - internal/app/service.go - Add validation to AddSecret, GetSecret, DeleteSecret
- **Why:** Prevents path traversal attacks and filesystem issues
- **Impact:** High - security

#### 1.3 Missing Features (11-missing-features.md) ❌ TODO
- **Current state:** `ListAll()` exists in service.go but returns nil
- **Files to modify:**
  - internal/storage/repository.go - Add ListFiles() method to interface
  - internal/storage/file/store.go - Implement file listing
  - internal/app/service.go - Implement ListAll() properly (or rename to ListSecrets)
  - Create internal/cli/list.go - Add list command
  - internal/cli/root.go - Register the list command
- **Impact:** Medium - feature completeness

### Phase 2: Architecture and Code Quality

**Goal:** Improve code structure and maintainability.

#### 2.1 Clean Architecture (00-architecture-improvements.md) ⚠️ PARTIAL
- ✅ Service interface (MSKService) is well-defined
- ✅ Repository interface exists
- ❌ **TODO:** Add PersistentPostRun cleanup hook in root.go
- ❌ **TODO:** Add ListFiles() to Repository interface (blocked by 1.3)

#### 2.2 Service Layer Design (06-service-layer-design.md) ✅ COMPLETED
- Clean separation of concerns is already in place
- Service layer properly delegates to repository and encryption

#### 2.3 Context Usage (05-context-usage.md) ✅ COMPLETED
- Context is properly checked in all storage layer methods
- Context cancellation pattern is correctly implemented

### Phase 3: Security Improvements (Non-Memory)

**Goal:** Fix security issues that do not require complex memory management.

#### 3.1 Argon2 Configuration (02-argon2-configuration.md) ✅ COMPLETED
- **What was done:** Memory increased from 64 (64KB) to 64*1024 (64MB)
- Parameters are now: time=3, memory=64MB, threads=4, keyLen=32
- ⚠️ **Note:** Existing vault files from old parameters will be incompatible

#### 3.2 Storage Security (07-storage-security.md) ✅ COMPLETED
- Atomic writes implemented using temp files + rename
- Directory sync for durability
- File permissions set to 0o600 (owner read/write only)

### Phase 4: Configuration and Polish

**Goal:** Make the project more flexible and user-friendly.

#### 4.1 Configuration System (10-configuration.md) ❌ TODO
- Vault path is still hardcoded as "./vault/"
- **Consider if this is really needed** - for a simple CLI tool, hardcoded paths are acceptable

#### 4.2 Dependency Management (08-dependency-management.md) ⚠️ REVIEW
- go.mod shows `go 1.25.5` - verify this is intentional
- Dependencies look reasonable

#### 4.3 Cobra CLI Improvements (09-cobra-cli-practices.md) ❌ TODO
- **Unused flags still present in root.go:**
  - `--master` / `-m` - defined but never used
  - `--toggle` / `-t` - defined but never used
- **Action:** Remove these flags or implement them

### Phase 5: Memory Safety (Do Last)

**Goal:** Implement secure memory management (most complex topic).

#### 5.1 Memory Safety Implementation (01-memory-safety.md) ⚠️ PARTIAL
- **What's done:**
  - ✅ `cleanupByte()` function exists in decrypt.go
  - ✅ Derived key is cleared after decryption: `defer cleanupByte(key)`
  - ✅ Plaintext is cleared after decryption: `defer cleanupByte(plaintext)`
- **What's missing:**
  - ❌ encrypt.go does NOT clear the derived key after use
  - ❌ Master key is never cleared from memory
  - ❌ No SecureBuffer implementation
  - ❌ No memory locking (mlock/VirtualLock)
  - ❌ No PersistentPostRun hook to clear master key on exit
  - ❌ Password prompt buffer not cleared after use

---

## Recommended Next Steps (Priority Order)

### Immediate (High Priority)
1. **1.2 Input Validation** - Security vulnerability, prevents path traversal
2. **1.3 Missing Features** - ListAll is a stub, implement properly
3. **4.3 Remove unused flags** - Quick win, improves UX

### Short Term (Medium Priority)
4. **5.1 Memory Safety (encrypt.go)** - Add `defer cleanupByte(key)` to Encrypt function
5. **2.1 Add cleanup hook** - Add PersistentPostRun to clear master key

### Optional (Lower Priority)
6. **4.1 Configuration** - Consider if really needed
7. **4.2 Go version** - Verify go.mod version is correct

---

## Quick Reference: What Each File Needs

| File | Status | Remaining Work |
|------|--------|----------------|
| cmd/msk/main.go | ✅ | None |
| internal/cli/root.go | ⚠️ | Remove unused flags, add cleanup hook |
| internal/cli/add.go | ✅ | None |
| internal/cli/get.go | ✅ | None |
| internal/cli/delete.go | ✅ | None |
| internal/cli/list.go | ❌ | Create file |
| internal/app/service.go | ⚠️ | Add validation, implement ListAll |
| internal/domain/validation.go | ❌ | Create file |
| internal/storage/repository.go | ⚠️ | Add ListFiles() to interface |
| internal/storage/file/store.go | ⚠️ | Implement ListFiles() |
| internal/encryption/key.go | ✅ | None |
| internal/encryption/encrypt.go | ⚠️ | Add defer cleanupByte(key) |
| internal/encryption/decrypt.go | ✅ | None |

---

## Notes

- Memory safety (Phase 5) is intentionally last because it is the most complex topic
- Each improvement file contains detailed implementation code and explanations
- The architecture improvements (Phase 2) prepare the codebase for easier memory safety integration later
- Focus on getting the project working well before tackling memory safety

This plan ensures you learn incrementally while building a production-ready password manager.
