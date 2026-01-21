# Group 10: Configuration and Flexibility

**Priority:** 🟡 **MEDIUM** - Usability (Phase 4)

## Current Problems

**File: `cmd/msk/main.go`**

```go
// Hardcoded vault path
store, _ := file.NewStore("./vault/")
```

**File: `internal/encryption/key.go`**

```go
// Hardcoded Argon2 parameters
return argon2.IDKey(password, salt, 3, 64, 4, 32)
```

## Issues

1. **Hardcoded vault path** - `./vault/` is relative to current directory
2. **No configuration file support** - Users can't customize behavior
3. **Hardcoded Argon2 parameters** - Can't adjust for different hardware
4. **No environment variable support** - Can't configure via env

## Learning Topics

- Configuration management patterns
- XDG Base Directory Specification
- Environment variables for configuration
- Configuration file formats (TOML, YAML, JSON)

## Implementation

### Step 1: Create Configuration Package

**Create: `internal/config/config.go`**

```go
package config

import (
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "runtime"
)

// Config holds all configurable settings for MSK
type Config struct {
    // VaultPath is the directory where encrypted secrets are stored
    VaultPath string `json:"vault_path"`
    
    // Argon2 parameters for key derivation
    Argon2 Argon2Config `json:"argon2"`
}

// Argon2Config holds Argon2 key derivation parameters
type Argon2Config struct {
    Time    uint32 `json:"time"`    // Number of iterations
    Memory  uint32 `json:"memory"`  // Memory in KB
    Threads uint8  `json:"threads"` // Parallel threads
    KeyLen  uint32 `json:"key_len"` // Output key length
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
    return &Config{
        VaultPath: defaultVaultPath(),
        Argon2: Argon2Config{
            Time:    2,
            Memory:  64 * 1024, // 64 MB
            Threads: 4,
            KeyLen:  32,
        },
    }
}

// Load loads configuration from the default location
// Priority: environment variables > config file > defaults
func Load() (*Config, error) {
    cfg := DefaultConfig()
    
    // Try to load config file
    configPath := getConfigPath()
    if _, err := os.Stat(configPath); err == nil {
        if err := cfg.loadFromFile(configPath); err != nil {
            return nil, fmt.Errorf("failed to load config: %w", err)
        }
    }
    
    // Override with environment variables
    cfg.loadFromEnv()
    
    return cfg, nil
}

func (c *Config) loadFromFile(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    
    return json.Unmarshal(data, c)
}

func (c *Config) loadFromEnv() {
    // Vault path from environment
    if vault := os.Getenv("MSK_VAULT_PATH"); vault != "" {
        c.VaultPath = vault
    }
    
    // Could add more env vars for Argon2 params if needed
    // MSK_ARGON2_MEMORY, MSK_ARGON2_TIME, etc.
}

// Save saves the current configuration to the default location
func (c *Config) Save() error {
    configPath := getConfigPath()
    
    // Ensure directory exists
    configDir := filepath.Dir(configPath)
    if err := os.MkdirAll(configDir, 0o700); err != nil {
        return err
    }
    
    data, err := json.MarshalIndent(c, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(configPath, data, 0o600)
}

// getConfigPath returns the path to the config file
func getConfigPath() string {
    // Check XDG_CONFIG_HOME first
    if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
        return filepath.Join(xdgConfig, "msk", "config.json")
    }
    
    home, err := os.UserHomeDir()
    if err != nil {
        return "msk.json" // Fallback to current directory
    }
    
    if runtime.GOOS == "windows" {
        // Windows: %APPDATA%\msk\config.json
        if appData := os.Getenv("APPDATA"); appData != "" {
            return filepath.Join(appData, "msk", "config.json")
        }
        return filepath.Join(home, "AppData", "Roaming", "msk", "config.json")
    }
    
    // Unix: ~/.config/msk/config.json
    return filepath.Join(home, ".config", "msk", "config.json")
}

// defaultVaultPath returns the default vault directory
func defaultVaultPath() string {
    // Check XDG_DATA_HOME first
    if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
        return filepath.Join(xdgData, "msk", "vault")
    }
    
    home, err := os.UserHomeDir()
    if err != nil {
        return "./vault" // Fallback
    }
    
    if runtime.GOOS == "windows" {
        // Windows: %LOCALAPPDATA%\msk\vault
        if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
            return filepath.Join(localAppData, "msk", "vault")
        }
        return filepath.Join(home, "AppData", "Local", "msk", "vault")
    }
    
    // Unix: ~/.local/share/msk/vault
    return filepath.Join(home, ".local", "share", "msk", "vault")
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
    if c.VaultPath == "" {
        return errors.New("vault_path cannot be empty")
    }
    
    if c.Argon2.Memory < 1024 { // Less than 1MB
        return errors.New("argon2 memory too low (minimum 1MB)")
    }
    
    if c.Argon2.Time < 1 {
        return errors.New("argon2 time must be at least 1")
    }
    
    return nil
}
```

### Step 2: Update Main to Use Config

**Update: `cmd/msk/main.go`**

```go
package main

import (
    "fmt"
    "os"

    "github.com/amauribechtoldjr/msk/internal/app"
    "github.com/amauribechtoldjr/msk/internal/cli"
    "github.com/amauribechtoldjr/msk/internal/clip"
    "github.com/amauribechtoldjr/msk/internal/config"
    "github.com/amauribechtoldjr/msk/internal/encryption"
    "github.com/amauribechtoldjr/msk/internal/logger"
    "github.com/amauribechtoldjr/msk/internal/storage/file"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        logger.RenderError(fmt.Errorf("failed to load config: %w", err))
        os.Exit(1)
    }
    
    if err := cfg.Validate(); err != nil {
        logger.RenderError(fmt.Errorf("invalid configuration: %w", err))
        os.Exit(1)
    }
    
    // Initialize clipboard
    if err := clip.Init(); err != nil {
        logger.RenderError(fmt.Errorf("clipboard init failed: %w", err))
        os.Exit(1)
    }
    
    // Initialize storage with configured path
    store, err := file.NewStore(cfg.VaultPath)
    if err != nil {
        logger.RenderError(fmt.Errorf("failed to initialize vault: %w", err))
        os.Exit(1)
    }
    
    // Initialize encryption with configured Argon2 params
    enc := encryption.NewArgonCrypt(encryption.ArgonConfig{
        Time:    cfg.Argon2.Time,
        Memory:  cfg.Argon2.Memory,
        Threads: cfg.Argon2.Threads,
        KeyLen:  cfg.Argon2.KeyLen,
    })
    
    service := app.NewMSKService(store, enc)
    
    rootCmd := cli.NewMSKCmd(service)
    if err := rootCmd.Execute(); err != nil {
        logger.RenderError(err)
        os.Exit(1)
    }
}
```

### Step 3: Update Encryption to Accept Config

**Update: `internal/encryption/encryption.go`**

```go
package encryption

// ArgonConfig holds Argon2 parameters
type ArgonConfig struct {
    Time    uint32
    Memory  uint32
    Threads uint8
    KeyLen  uint32
}

// DefaultArgonConfig returns secure defaults
func DefaultArgonConfig() ArgonConfig {
    return ArgonConfig{
        Time:    2,
        Memory:  64 * 1024, // 64 MB
        Threads: 4,
        KeyLen:  32,
    }
}

type ArgonCrypt struct {
    mk     []byte
    config ArgonConfig
}

func NewArgonCrypt(cfg ArgonConfig) *ArgonCrypt {
    return &ArgonCrypt{config: cfg}
}

// In key.go, use the config:
func (a *ArgonCrypt) deriveKey(password, salt []byte) []byte {
    return argon2.IDKey(
        password,
        salt,
        a.config.Time,
        a.config.Memory,
        a.config.Threads,
        a.config.KeyLen,
    )
}
```

## Configuration File Example

**~/.config/msk/config.json**

```json
{
  "vault_path": "/home/user/.local/share/msk/vault",
  "argon2": {
    "time": 2,
    "memory": 65536,
    "threads": 4,
    "key_len": 32
  }
}
```

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `MSK_VAULT_PATH` | Custom vault directory | `/path/to/vault` |
| `XDG_CONFIG_HOME` | Config directory (Unix) | `~/.config` |
| `XDG_DATA_HOME` | Data directory (Unix) | `~/.local/share` |

## Pattern to Learn

**Configuration Hierarchy:**
1. Command-line flags (highest priority)
2. Environment variables
3. Configuration file
4. Default values (lowest priority)

**XDG Base Directory Specification:**
- Config files: `$XDG_CONFIG_HOME/app/` (default: `~/.config/app/`)
- Data files: `$XDG_DATA_HOME/app/` (default: `~/.local/share/app/`)
- Cache files: `$XDG_CACHE_HOME/app/` (default: `~/.cache/app/`)

**Platform Conventions:**
- **Linux/macOS**: Follow XDG specification
- **Windows**: Use `%APPDATA%` for config, `%LOCALAPPDATA%` for data
