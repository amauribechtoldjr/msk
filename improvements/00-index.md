# Recommended Learning Order

This guide organizes improvements so you build a solid foundation first, leaving memory safety (the most complex topic) for last.

## Phase 1: Foundation & Reliability (Do First!)

1. **[Error Handling](03-error-handling.md)** 🔴 **START HERE**
   - Fix ignored errors in main.go and add.go
   - Ensures the app fails gracefully

2. **[Input Validation](04-input-validation.md)** 🔴 **CRITICAL**
   - Prevent path traversal attacks
   - Add length limits and character restrictions

3. **[Missing Features](11-missing-features.md)** 🟡 **USEFUL**
   - Implement ListAll() functionality
   - Complete core features

## Phase 2: Architecture & Code Quality

4. **[Architecture Improvements](00-architecture-improvements.md)** 🟡 **IMPORTANT**
   - Clean architecture principles
   - Layer responsibilities and separation of concerns
   - Interface design patterns

5. **[Service Layer Design](06-service-layer-design.md)** 🟡 **MEDIUM**
   - Improve interface design
   - Ensure proper separation of concerns

6. **[Context Usage](05-context-usage.md)** 🟡 **MEDIUM**
   - Add cancellation support for long-running operations

## Phase 3: Security Improvements (Non-Memory)

7. **[Argon2 Configuration](02-argon2-configuration.md)** 🟠 **HIGH**
   - Fix weak Argon2 parameters (64KB → 64MB)
   - Critical security vulnerability

8. **[Storage Security](07-storage-security.md)** 🟡 **MEDIUM**
   - Add file locking for concurrent access
   - Improve atomic write guarantees

## Phase 4: Configuration & Polish

9. **[Configuration System](10-configuration.md)** 🟡 **MEDIUM**
   - Make vault path configurable
   - Externalize Argon2 parameters

10. **[Dependency Management](08-dependency-management.md)** 🟢 **LOW**
    - Fix go.mod issues
    - Quick fix

11. **[Cobra CLI Improvements](09-cobra-cli-practices.md)** 🟢 **LOW**
    - Remove unused flags
    - Improve documentation

## Phase 5: Memory Safety (Do Last!)

12. **[Memory Safety](01-memory-safety.md)** 🔴 **COMPLEX** - Save for last!
    - SecureBuffer implementation
    - Platform-specific memory locking (mlock/VirtualLock)
    - Secure memory zeroing
    - Clearing passwords and keys from memory

---

## Why This Order?

1. **Foundation First**: Error handling and input validation ensure the app works reliably
2. **Architecture Second**: Clean architecture makes future changes easier
3. **Security Third**: Fix obvious security issues (Argon2) that are straightforward
4. **Polish Fourth**: Configuration and cleanup improve user experience
5. **Memory Safety Last**: Most complex topic, requires understanding of:
   - Go memory model
   - Platform-specific system calls
   - Compiler optimizations
   - Ownership and lifecycle management

## Execution Plan

For a detailed step-by-step guide, see **[execution-plan.md](execution-plan.md)**.

---

*Build the foundation first. Memory safety is easier to add when the architecture is clean.*
