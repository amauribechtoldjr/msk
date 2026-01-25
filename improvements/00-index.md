# Recommended Learning Order

This guide organizes improvements so you build a solid foundation first, leaving memory safety (the most complex topic) for last.

---

## Current Progress

| Phase | Items Done | Items Remaining |
|-------|------------|-----------------|
| Phase 1 | 1/3 | Input validation, List feature |
| Phase 2 | 2/3 | Cleanup hooks |
| Phase 3 | 2/2 | ✅ Complete |
| Phase 4 | 0/3 | Config, deps, CLI cleanup |
| Phase 5 | 0.5/1 | Memory safety (partial) |

---

## Phase 1: Foundation & Reliability

1. **[Error Handling](03-error-handling.md)** ✅ **DONE**
   - ~~Fix ignored errors in main.go and add.go~~
   - ~~Ensures the app fails gracefully~~

2. **[Input Validation](04-input-validation.md)** 🔴 **TODO - HIGH PRIORITY**
   - Prevent path traversal attacks
   - Add length limits and character restrictions

3. **[Missing Features](11-missing-features.md)** 🔴 **TODO**
   - Implement ListAll() functionality (currently returns nil)
   - Add list CLI command
   - Complete core features

## Phase 2: Architecture & Code Quality

4. **[Architecture Improvements](00-architecture-improvements.md)** ⚠️ **PARTIAL**
   - ✅ Clean architecture principles applied
   - ✅ Layer responsibilities and separation of concerns
   - ❌ Need cleanup hooks (PersistentPostRun)

5. **[Service Layer Design](06-service-layer-design.md)** ✅ **DONE**
   - ~~Improve interface design~~
   - ~~Ensure proper separation of concerns~~

6. **[Context Usage](05-context-usage.md)** ✅ **DONE**
   - ~~Context properly checked in storage layer~~

## Phase 3: Security Improvements (Non-Memory)

7. **[Argon2 Configuration](02-argon2-configuration.md)** ✅ **DONE**
   - ~~Fix weak Argon2 parameters (64KB → 64MB)~~
   - Memory now at 64*1024 (64MB)

8. **[Storage Security](07-storage-security.md)** ✅ **DONE**
   - ~~Atomic writes with temp file + rename~~
   - ~~Directory sync for durability~~

## Phase 4: Configuration & Polish

9. **[Configuration System](10-configuration.md)** 🟡 **OPTIONAL**
   - Make vault path configurable
   - Consider if really needed for a simple CLI tool

10. **[Dependency Management](08-dependency-management.md)** 🟡 **REVIEW**
    - Check go.mod Go version (1.25.5)

11. **[Cobra CLI Improvements](09-cobra-cli-practices.md)** 🟡 **TODO**
    - Remove unused flags (--toggle, --master)
    - Add list command

## Phase 5: Memory Safety (Do Last!)

12. **[Memory Safety](01-memory-safety.md)** ⚠️ **PARTIAL**
    - ✅ cleanupByte() function exists
    - ✅ Derived key cleared in decrypt.go
    - ✅ Plaintext cleared in decrypt.go
    - ❌ encrypt.go does not clear derived key
    - ❌ Master key never cleared
    - ❌ No SecureBuffer implementation
    - ❌ No memory locking

---

## Next Steps (Priority Order)

1. **Input Validation** (04-input-validation.md) - Security critical
2. **List Feature** (11-missing-features.md) - Complete core functionality
3. **Remove unused CLI flags** (09-cobra-cli-practices.md) - Quick win
4. **Add cleanup to encrypt.go** - Match what decrypt.go does

---

## Execution Plan

For a detailed step-by-step guide, see **[execution-plan.md](execution-plan.md)**.

---

*Build the foundation first. Memory safety is easier to add when the architecture is clean.*
