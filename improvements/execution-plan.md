# MSK Improvements Execution Plan

This plan organizes all improvements from the `improvements/` folder into a logical execution order. The goal is to ensure the project has a solid architecture and runs correctly before tackling complex memory safety topics.

**Key Principle:** Memory safety (Phase 5) is intentionally last because it requires understanding Go's memory model, platform-specific system calls, and compiler optimizations.

## Execution Phases

### Phase 1: Foundation and Reliability (Do First)

**Goal:** Ensure the project runs correctly and handles errors properly.

#### 1.1 Error Handling (03-error-handling.md)
- **Files to modify:**
  - cmd/msk/main.go - Handle store creation error (line 41)
  - internal/cli/add.go - Handle flag parsing error (line 27)
- **Why first:** Prevents silent failures and ensures the app fails gracefully
- **Impact:** High - affects reliability

#### 1.2 Input Validation (04-input-validation.md)
- **Files to create:**
  - internal/domain/validation.go - Validation functions
- **Files to modify:**
  - internal/app/service.go - Add validation to AddSecret, GetSecret, DeleteSecret
- **Why early:** Prevents security vulnerabilities (path traversal) and filesystem issues
- **Impact:** High - security and reliability

#### 1.3 Missing Features (11-missing-features.md)
- **Files to modify:**
  - internal/app/service.go - Implement ListAll() method
  - internal/storage/repository.go - Add ListFiles() method
  - internal/storage/file/store.go - Implement file listing
  - internal/cli/root.go or create internal/cli/list.go - Add list command
- **Why early:** Completes core functionality
- **Impact:** Medium - feature completeness

### Phase 2: Architecture and Code Quality

**Goal:** Improve code structure and maintainability.

#### 2.1 Clean Architecture (00-architecture-improvements.md)
- **Files to review/modify:**
  - internal/app/service.go - Review interface design
  - internal/cli/root.go - Add PersistentPostRun cleanup hook
  - internal/storage/repository.go - Add ListFiles() to interface
- **Why now:** Establishes clean architecture patterns for future changes
- **Impact:** Medium - maintainability

#### 2.2 Service Layer Design (06-service-layer-design.md)
- **Files to modify:**
  - internal/app/service.go - Ensure proper separation of concerns
- **Why now:** Reinforces clean architecture
- **Impact:** Medium - maintainability

#### 2.3 Context Usage (05-context-usage.md)
- **Files to modify:**
  - internal/app/service.go - Add proper context cancellation support
  - internal/encryption/encrypt.go - Consider context for long Argon2 operations
- **Why now:** Improves cancellation support
- **Impact:** Medium - user experience

### Phase 3: Security Improvements (Non-Memory)

**Goal:** Fix security issues that do not require complex memory management.

#### 3.1 Argon2 Configuration (02-argon2-configuration.md)
- **Files to modify:**
  - internal/encryption/key.go - Update Argon2 parameters (memory from 64KB to 64MB)
- **Why now:** Critical security vulnerability, straightforward fix
- **Impact:** High - cryptographic security

#### 3.2 Storage Security (07-storage-security.md)
- **Files to modify:**
  - internal/storage/file/store.go - Add file locking for concurrent access
- **Why now:** Prevents data corruption
- **Impact:** Medium - data integrity

### Phase 4: Configuration and Polish

**Goal:** Make the project more flexible and user-friendly.

#### 4.1 Configuration System (10-configuration.md)
- **Files to create:**
  - internal/config/config.go - Configuration management
- **Files to modify:**
  - cmd/msk/main.go - Load configuration
  - internal/storage/file/store.go - Use configurable vault path
  - internal/encryption/key.go - Use configurable Argon2 parameters
- **Why now:** Makes the project more flexible
- **Impact:** Medium - flexibility

#### 4.2 Dependency Management (08-dependency-management.md)
- **Files to modify:**
  - go.mod - Fix Go version, mark dependencies correctly
- **Why now:** Quick fix
- **Impact:** Low - project hygiene

#### 4.3 Cobra CLI Improvements (09-cobra-cli-practices.md)
- **Files to modify:**
  - internal/cli/root.go - Remove unused flags (--toggle, --master)
- **Why now:** Clean up unused code
- **Impact:** Low - code cleanliness

### Phase 5: Memory Safety (Do Last)

**Goal:** Implement secure memory management (most complex topic).

#### 5.1 Memory Safety Implementation (01-memory-safety.md)
- **Files to create:**
  - internal/secure/secure.go - SecureBuffer implementation
  - internal/secure/lock_unix.go - Unix memory locking (mlock)
  - internal/secure/lock_windows.go - Windows memory locking (VirtualLock)
- **Files to modify:**
  - internal/encryption/encryption.go - Update to use SecureBuffer for master key
  - internal/encryption/encrypt.go - Clear derived keys and plaintext after use
  - internal/encryption/decrypt.go - Clear derived keys and plaintext after use
  - internal/app/service.go - Update to use SecureBuffer
  - internal/cli/root.go - Manage SecureBuffer lifecycle with cleanup hooks
  - internal/cli/add.go - Clear password buffers
  - internal/cli/get.go - Clear password buffers
- **Why last:** Requires understanding of:
  - Go memory model and garbage collector
  - Platform-specific system calls (mlock, VirtualLock)
  - Compiler optimization prevention (runtime.KeepAlive)
  - Ownership and lifecycle management
- **Impact:** High - security, but most complex

## Implementation Strategy

### For Each Phase
- Start with one improvement at a time - Do not try to do everything at once
- Test after each change - Ensure the app still works
- Commit after each improvement - Makes it easier to roll back if needed
- Read the improvement file - Each file has detailed implementation guidance

### Testing Approach
- Phases 1-4: Standard Go testing (unit tests, integration tests)
- Phase 5: Use memory analysis tools (e.g., Process Hacker on Windows) to verify memory clearing

## Key Principles

1. Foundation First: Build a solid base before adding complex features
2. Functionality Before Optimization: Ensure everything works before optimizing security
3. Incremental Learning: Each phase builds on the previous one
4. Test as You Go: Do not accumulate technical debt

## Estimated Complexity

- Phase 1: Low-Medium complexity (straightforward fixes)
- Phase 2: Medium complexity (requires understanding Go patterns)
- Phase 3: Low-Medium complexity (mostly configuration changes)
- Phase 4: Low complexity (configuration and cleanup)
- Phase 5: High complexity (memory management, platform-specific code)

## Notes

- Memory safety (Phase 5) is intentionally last because it is the most complex topic
- Each improvement file contains detailed implementation code and explanations
- You can skip phases if certain improvements are not needed
- The architecture improvements (Phase 2) prepare the codebase for easier memory safety integration later
- Focus on getting the project working well before tackling memory safety

This plan ensures you learn incrementally while building a production-ready password manager.
