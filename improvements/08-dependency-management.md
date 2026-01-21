# Group 8: Dependency Management and Module Configuration

**Priority:** 🟢 **LOW** - Code quality (Phase 4)

## Current Problems

**File: `go.mod`**

```go
// Current go.mod might have issues:
module github.com/amauribechtoldjr/msk

go 1.25.5  // This version doesn't exist!

require (
    github.com/spf13/cobra v1.8.0
    golang.org/x/crypto v0.21.0
    // ... some marked as indirect but should be direct
)
```

## Issues

1. **Invalid Go version** - 1.25.5 doesn't exist (likely a typo, should be 1.21.5 or 1.22.x)
2. **Dependencies marked incorrectly** - Some indirect deps should be direct
3. **Potential outdated dependencies** - Security updates may be available

## Learning Topics

- Go modules system
- Direct vs indirect dependencies
- Semantic versioning in Go
- `go mod tidy` and `go get -u`

## Implementation

### Step 1: Fix Go Version

Check your actual Go version:

```bash
go version
# Output: go version go1.22.0 windows/amd64
```

**Update: `go.mod`**

```go
module github.com/amauribechtoldjr/msk

go 1.22

require (
    github.com/atotto/clipboard v0.1.4
    github.com/spf13/cobra v1.8.1
    golang.org/x/crypto v0.28.0
    golang.org/x/term v0.25.0
)

require (
    github.com/inconshreveable/mousetrap v1.1.0 // indirect
    github.com/spf13/pflag v1.0.5 // indirect
    golang.org/x/sys v0.26.0 // indirect
)
```

### Step 2: Clean Up Dependencies

Run these commands to fix your go.mod:

```bash
# Update Go version in go.mod to match your installation
go mod edit -go=1.22

# Remove unused dependencies and add missing ones
go mod tidy

# Update all dependencies to latest minor/patch versions
go get -u ./...

# Tidy again after updates
go mod tidy

# Verify everything works
go build ./...
```

### Step 3: Check for Security Updates

```bash
# Check for known vulnerabilities
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## Understanding Direct vs Indirect

**Direct dependencies**: Packages you import directly in your code

```go
import (
    "github.com/spf13/cobra"        // Direct - you use this
    "golang.org/x/crypto/argon2"    // Direct - you use this
)
```

**Indirect dependencies**: Packages your dependencies depend on

```go
// You never import these, but cobra needs them:
// github.com/inconshreveable/mousetrap  // indirect
// github.com/spf13/pflag                 // indirect
```

## Best Practices

### 1. Pin Major Versions

```go
require (
    github.com/spf13/cobra v1.8.1  // Good: specific version
)
```

### 2. Update Regularly

```bash
# Check for updates
go list -m -u all

# Update specific package
go get -u github.com/spf13/cobra

# Update all
go get -u ./...
```

### 3. Use go.sum for Verification

The `go.sum` file contains cryptographic hashes. Never edit it manually:

```bash
# Regenerate if corrupted
rm go.sum
go mod tidy
```

### 4. Vendor Dependencies (Optional)

For reproducible builds:

```bash
go mod vendor
# Then build with: go build -mod=vendor ./...
```

## Recommended Dependencies for MSK

```go
module github.com/amauribechtoldjr/msk

go 1.22

require (
    // CLI framework
    github.com/spf13/cobra v1.8.1
    
    // Cryptography
    golang.org/x/crypto v0.28.0
    
    // Clipboard access
    github.com/atotto/clipboard v0.1.4
    
    // Terminal password input
    golang.org/x/term v0.25.0
    
    // Optional: File locking (if implementing storage security)
    // github.com/gofrs/flock v0.8.1
)
```

## Pattern to Learn

**Go Module Best Practices:**
- Run `go mod tidy` after adding/removing imports
- Keep Go version in go.mod up to date
- Update dependencies regularly for security fixes
- Use `govulncheck` to find vulnerabilities
- Understand the difference between direct and indirect dependencies
