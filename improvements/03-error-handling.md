# Group 3: Go Error Handling and Defensive Programming

**Priority:** 🟠 **HIGH** - Code reliability (Phase 1)

## Current Problems

### Problem 1: Ignored error in main.go (line 41)

```go
// CURRENT (BAD)
store, _ := file.NewStore("./vault/")  // Error ignored!
```

If the vault directory can't be created (permissions, disk full, etc.), the app continues with a nil store and will crash later with a confusing error.

### Problem 2: Ignored error in add.go (line 27)

```go
// CURRENT (BAD)
value, _ := cmd.Flags().GetString("password")  // Error ignored!
```

While Cobra flag parsing rarely fails, ignoring errors is a bad habit.

## Issues

1. **Ignored errors in `cmd/msk/main.go`** - Store creation error not handled
2. **Ignored errors in `internal/cli/add.go`** - Flag parsing error not handled
3. **Error handling patterns inconsistent** - Some places wrap, some don't

## Learning Topics

- Go error handling philosophy ("errors are values")
- Defensive programming - always check errors
- Error wrapping with `fmt.Errorf()` and `%w`
- When to handle vs propagate errors
- Custom error types for better error handling

## Implementation

### Fix 1: main.go

**Update: `cmd/msk/main.go`**

```go
package main

import (
    "fmt"
    "os"

    "github.com/amauribechtoldjr/msk/internal/app"
    "github.com/amauribechtoldjr/msk/internal/cli"
    "github.com/amauribechtoldjr/msk/internal/clip"
    "github.com/amauribechtoldjr/msk/internal/encryption"
    "github.com/amauribechtoldjr/msk/internal/logger"
    "github.com/amauribechtoldjr/msk/internal/storage/file"
)

func main() {
    if err := clip.Init(); err != nil {
        logger.RenderError(fmt.Errorf("clipboard initialization failed: %w", err))
        os.Exit(1)
    }

    store, err := file.NewStore("./vault/")
    if err != nil {
        logger.RenderError(fmt.Errorf("failed to initialize vault: %w", err))
        os.Exit(1)
    }

    enc := encryption.NewArgonCrypt()
    service := app.NewMSKService(store, enc)

    rootCmd := cli.NewMSKCmd(service)
    if err := rootCmd.Execute(); err != nil {
        logger.RenderError(err)
        os.Exit(1)
    }
}
```

### Fix 2: add.go

**Update: `internal/cli/add.go`**

```go
func NewAddCmd(service app.MSKService) *cobra.Command {
    addCmd := &cobra.Command{
        Use:     "add <name>",
        Aliases: []string{"a"},
        Short:   "Add a password to the vault",
        Long: `Add a new password entry to the vault.

You can provide the password via the -p flag or enter it interactively.

Examples:
  msk add github
  msk add github -p "my-secure-password"`,
        SilenceErrors: true,
        SilenceUsage:  true,
        Args:          cobra.ExactArgs(1), // Use Cobra's built-in validation
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()
            name := args[0]

            // Handle flag parsing error
            value, err := cmd.Flags().GetString("password")
            if err != nil {
                return fmt.Errorf("failed to read password flag: %w", err)
            }
            
            password := []byte(value)

            if value == "" {
                password, err = PromptPassword("Enter password:")
                if err != nil {
                    return fmt.Errorf("failed to read password: %w", err)
                }
            }

            if err := service.AddSecret(ctx, name, password); err != nil {
                return fmt.Errorf("failed to add secret: %w", err)
            }

            logger.PrintSuccess("Password added successfully\n")
            return nil
        },
    }

    addCmd.Flags().StringP("password", "p", "", "Password to store (interactive if not provided)")

    return addCmd
}
```

### Fix 3: get.go (check for similar issues)

**Update: `internal/cli/get.go`**

```go
func NewGetCmd(service app.MSKService) *cobra.Command {
    getCmd := &cobra.Command{
        Use:     "get <name>",
        Aliases: []string{"g"},
        Short:   "Get a password from the vault",
        Long: `Retrieve a password from the vault and copy it to the clipboard.

The password will be available for pasting until you copy something else.

Examples:
  msk get github
  msk get my-email-password`,
        SilenceErrors: true,
        SilenceUsage:  true,
        Args:          cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()
            name := args[0]

            password, err := service.GetSecret(ctx, name)
            if err != nil {
                return fmt.Errorf("failed to retrieve secret: %w", err)
            }

            if err := clip.Write(password); err != nil {
                return fmt.Errorf("failed to copy to clipboard: %w", err)
            }

            logger.PrintSuccess("Password copied to clipboard\n")
            return nil
        },
    }

    return getCmd
}
```

## Error Wrapping Best Practices

### When to Wrap Errors

```go
// DO wrap when adding context
if err := doSomething(); err != nil {
    return fmt.Errorf("failed to do something for user %s: %w", userID, err)
}

// DON'T wrap when no additional context
if err := doSomething(); err != nil {
    return err  // Just propagate
}
```

### Using errors.Is and errors.As

```go
import "errors"

// Check for specific error types
if errors.Is(err, file.ErrNotFound) {
    // Handle not found case
}

// Unwrap to get underlying error
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    fmt.Println("Path:", pathErr.Path)
}
```

## Pattern to Learn

**Go Error Philosophy:**
- Errors are values - treat them as first-class citizens
- Always handle or explicitly propagate errors
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Use `errors.Is` and `errors.As` for error checking
- If you must ignore an error, document why: `_ = file.Close() // Best effort, error already handled`

**Defensive Programming:**
- Assume external operations can fail (I/O, network, parsing)
- Fail fast with clear error messages
- Don't let errors propagate silently to cause confusing crashes later
