# MSK Code Review - Improvement Plan

This document serves as an index to a comprehensive code review summary, organized by learning topics with security and Go design patterns as priorities.

Each improvement group has been separated into its own file in the `improvements/` directory for better readability and organization.

## Quick Navigation

- **[Recommended Learning Order](improvements/00-index.md)** - Suggested approach for implementing improvements
- **[Key Principles](improvements/principles.md)** - Security principles and Go design patterns
- **[Testing Guide](improvements/testing.md)** - How to test your improvements
- **[Resources](improvements/resources.md)** - Additional learning resources

## Improvement Groups

1. **[Group 1: Critical Security - Memory Safety](improvements/01-memory-safety.md)** 🔴 **CRITICAL**
   - Master key never cleared from memory
   - Passwords stored in memory but not cleared after use
   - Derived encryption keys not cleared after encryption

2. **[Group 2: Cryptographic Security - Argon2 Configuration](improvements/02-argon2-configuration.md)** 🟠 **HIGH**
   - Weak Argon2 parameters (time=3, memory=64KB is extremely low)
   - Using `argon2.IDKey` instead of `argon2id`

3. **[Group 3: Go Error Handling & Defensive Programming](improvements/03-error-handling.md)** 🟠 **HIGH**
   - Ignored errors in `main.go` (store creation), `add.go` (flag parsing)
   - Error handling patterns inconsistent

4. **[Group 4: Input Validation & Sanitization](improvements/04-input-validation.md)** 🟡 **MEDIUM-HIGH**
   - No input validation on secret names (path traversal risk: `../`, `/`)
   - No length limits on passwords or names

5. **[Group 5: Context Usage & Cancellation](improvements/05-context-usage.md)** 🟡 **MEDIUM**
   - Context passed but not fully utilized
   - No cancellation support for long-running operations (Argon2)

6. **[Group 6: Service Layer Design & Separation of Concerns](improvements/06-service-layer-design.md)** 🟡 **MEDIUM**
   - Service layer leaks implementation (returns `[]byte` instead of domain types)
   - Mixed concerns (service layer shouldn't know about clipboard)

7. **[Group 7: Storage & File System Security](improvements/07-storage-security.md)** 🟡 **MEDIUM**
   - No file locking (concurrent access risk)
   - Atomic writes not guaranteed on all platforms (Windows rename behavior)

8. **[Group 8: Dependency Management & Module Configuration](improvements/08-dependency-management.md)** 🟢 **LOW**
   - Dependencies marked as `// indirect` but should be direct
   - Go version in `go.mod` seems incorrect (1.25.5 doesn't exist)

9. **[Group 9: Cobra CLI Best Practices](improvements/09-cobra-cli-practices.md)** 🟢 **LOW**
   - Unused flags (`--toggle`, `--master` flag defined but not read)
   - Missing command documentation (empty `Long` descriptions)

10. **[Group 10: Configuration & Flexibility](improvements/10-configuration.md)** 🟡 **MEDIUM**
    - Hardcoded vault path (`./vault/`)
    - No configuration system
    - Argon2 parameters hardcoded

11. **[Group 11: Missing Features & Incomplete Implementation](improvements/11-missing-features.md)** 🟡 **LOW-MEDIUM**
    - `ListAll()` not implemented (returns `nil`)
    - Missing unit tests (entire codebase)

---

*Last Updated: Based on code review conducted on the MSK password manager codebase.*
