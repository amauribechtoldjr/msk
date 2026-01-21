# Group 9: Cobra CLI Best Practices

**Priority:** 🟢 **LOW** - UX and maintainability (Phase 4)

## Current Problems

**File: `internal/cli/root.go`**

```go
// Unused flags
cmd.PersistentFlags().StringP("master", "m", "", "Set the master key manually.")
cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
```

These flags are defined but never read, confusing users who see them in `--help`.

## Issues

1. **Unused flags** - `--toggle` and `--master` are defined but not used
2. **Empty Long descriptions** - Some commands have empty Long text
3. **Manual argument validation** - Using `len(args) < 1` instead of Cobra's built-in
4. **Missing examples** - Help output doesn't show usage examples

## Learning Topics

- Cobra framework best practices
- CLI UX design principles
- Command documentation
- Built-in argument validation

## Implementation

### Step 1: Remove Unused Flags

**Update: `internal/cli/root.go`**

```go
package cli

import (
    "github.com/amauribechtoldjr/msk/internal/app"
    "github.com/spf13/cobra"
)

func NewMSKCmd(service app.MSKService) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "msk",
        Short: "A secure, offline password manager",
        Long: `MSK is a lightweight password manager that stores credentials
locally on your computer. All passwords are encrypted using a 
master password with Argon2 key derivation and AES-256-GCM encryption.

Features:
  - Offline storage (no cloud, no network)
  - Strong encryption (Argon2 + AES-256-GCM)
  - Clipboard integration
  - Simple CLI interface

Get started:
  msk add github        # Add a new password
  msk get github        # Retrieve and copy to clipboard
  msk list              # Show all stored passwords
  msk delete github     # Remove a password`,
        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            // Skip for help commands
            if cmd.Name() == "help" || cmd.Name() == "completion" {
                return nil
            }
            
            ctx := cmd.Context()
            mk, err := PromptPassword("Enter master password:")
            if err != nil {
                return err
            }
            
            service.ConfigMK(ctx, mk)
            return nil
        },
        PersistentPostRun: func(cmd *cobra.Command, args []string) {
            // Future: Clear master key from memory
            // service.ClearMasterKey()
        },
    }
    
    // REMOVED: Unused flags
    // cmd.PersistentFlags().StringP("master", "m", "", "...")  // REMOVED
    // cmd.Flags().BoolP("toggle", "t", false, "...")           // REMOVED
    
    // Add subcommands
    cmd.AddCommand(NewAddCmd(service))
    cmd.AddCommand(NewGetCmd(service))
    cmd.AddCommand(NewDeleteCmd(service))
    cmd.AddCommand(NewListCmd(service))  // New!
    
    return cmd
}
```

### Step 2: Improve Add Command

**Update: `internal/cli/add.go`**

```go
package cli

import (
    "fmt"

    "github.com/amauribechtoldjr/msk/internal/app"
    "github.com/amauribechtoldjr/msk/internal/logger"
    "github.com/spf13/cobra"
)

func NewAddCmd(service app.MSKService) *cobra.Command {
    cmd := &cobra.Command{
        Use:     "add <name>",
        Aliases: []string{"a", "new"},
        Short:   "Add a new password to the vault",
        Long: `Add a new password entry to the vault.

The password can be provided via the -p flag or entered interactively
(recommended for security - interactive input is not stored in shell history).`,
        Example: `  # Add interactively (recommended)
  msk add github
  
  # Add with password flag (less secure - visible in shell history)
  msk add github -p "my-password"
  
  # Using alias
  msk a github`,
        Args:          cobra.ExactArgs(1),  // Built-in validation!
        SilenceErrors: true,
        SilenceUsage:  true,
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()
            name := args[0]  // Guaranteed to exist due to ExactArgs(1)

            value, err := cmd.Flags().GetString("password")
            if err != nil {
                return fmt.Errorf("failed to read password flag: %w", err)
            }
            
            password := []byte(value)
            if value == "" {
                password, err = PromptPassword("Enter password:")
                if err != nil {
                    return err
                }
            }

            if err := service.AddSecret(ctx, name, password); err != nil {
                return err
            }

            logger.PrintSuccess("Password '%s' added successfully\n", name)
            return nil
        },
    }

    cmd.Flags().StringP("password", "p", "", "Password to store (omit for interactive input)")

    return cmd
}
```

### Step 3: Improve Get Command

**Update: `internal/cli/get.go`**

```go
package cli

import (
    "fmt"

    "github.com/amauribechtoldjr/msk/internal/app"
    "github.com/amauribechtoldjr/msk/internal/clip"
    "github.com/amauribechtoldjr/msk/internal/logger"
    "github.com/spf13/cobra"
)

func NewGetCmd(service app.MSKService) *cobra.Command {
    cmd := &cobra.Command{
        Use:     "get <name>",
        Aliases: []string{"g"},
        Short:   "Retrieve a password and copy to clipboard",
        Long: `Retrieve a password from the vault and copy it to the clipboard.

The password will be available for pasting (Ctrl+V) until you copy 
something else. The password is never displayed on screen.`,
        Example: `  msk get github
  msk g my-email`,
        Args:          cobra.ExactArgs(1),
        SilenceErrors: true,
        SilenceUsage:  true,
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()
            name := args[0]

            password, err := service.GetSecret(ctx, name)
            if err != nil {
                return fmt.Errorf("failed to get '%s': %w", name, err)
            }

            if err := clip.Write(password); err != nil {
                return fmt.Errorf("failed to copy to clipboard: %w", err)
            }

            logger.PrintSuccess("Password '%s' copied to clipboard\n", name)
            return nil
        },
    }

    return cmd
}
```

### Step 4: Improve Delete Command

**Update: `internal/cli/delete.go`**

```go
package cli

import (
    "fmt"

    "github.com/amauribechtoldjr/msk/internal/app"
    "github.com/amauribechtoldjr/msk/internal/logger"
    "github.com/spf13/cobra"
)

func NewDeleteCmd(service app.MSKService) *cobra.Command {
    cmd := &cobra.Command{
        Use:     "delete <name>",
        Aliases: []string{"d", "rm", "remove"},
        Short:   "Delete a password from the vault",
        Long: `Permanently delete a password entry from the vault.

This action cannot be undone!`,
        Example: `  msk delete github
  msk rm old-password`,
        Args:          cobra.ExactArgs(1),
        SilenceErrors: true,
        SilenceUsage:  true,
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()
            name := args[0]

            if err := service.DeleteSecret(ctx, name); err != nil {
                return fmt.Errorf("failed to delete '%s': %w", name, err)
            }

            logger.PrintSuccess("Password '%s' deleted\n", name)
            return nil
        },
    }

    return cmd
}
```

### Step 5: Add List Command

**Create: `internal/cli/list.go`**

```go
package cli

import (
    "fmt"

    "github.com/amauribechtoldjr/msk/internal/app"
    "github.com/spf13/cobra"
)

func NewListCmd(service app.MSKService) *cobra.Command {
    cmd := &cobra.Command{
        Use:     "list",
        Aliases: []string{"ls", "l"},
        Short:   "List all stored password names",
        Long: `List all password entries in the vault.

Only the names are shown, not the actual passwords.`,
        Example: `  msk list
  msk ls`,
        Args:          cobra.NoArgs,
        SilenceErrors: true,
        SilenceUsage:  true,
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()

            names, err := service.ListSecrets(ctx)
            if err != nil {
                return fmt.Errorf("failed to list secrets: %w", err)
            }

            if len(names) == 0 {
                fmt.Println("No passwords stored yet.")
                fmt.Println("Use 'msk add <name>' to add one.")
                return nil
            }

            fmt.Printf("Stored passwords (%d):\n", len(names))
            for _, name := range names {
                fmt.Printf("  - %s\n", name)
            }

            return nil
        },
    }

    return cmd
}
```

## Cobra Best Practices Summary

| Practice | Bad | Good |
|----------|-----|------|
| Args validation | `if len(args) < 1` | `Args: cobra.ExactArgs(1)` |
| Documentation | Empty `Long` | Descriptive with examples |
| Unused flags | Define and never use | Remove or implement |
| Aliases | None | Provide short alternatives |
| Examples | None | Show common use cases |

## Pattern to Learn

**CLI UX Principles:**
- Use built-in Cobra features (Args, aliases, examples)
- Provide clear, helpful error messages
- Show examples in help text
- Remove unused code that confuses users
- Use consistent naming across commands
