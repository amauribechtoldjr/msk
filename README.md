# MSK

**A lightweight, offline password manager that keeps your credentials encrypted and local.**

![Go Version](https://img.shields.io/github/go-mod/go-version/amauribechtoldjr/msk)
![License](https://img.shields.io/github/license/amauribechtoldjr/msk)
![CI](https://img.shields.io/github/actions/workflow/status/amauribechtoldjr/msk/ci.yml?label=CI)

MSK stores all your passwords locally on your machine, encrypted with a master password. No cloud, no sync, no network requests — your credentials never leave your computer.

## Features

- **AES-256-GCM encryption** — Every secret is individually encrypted at rest
- **Argon2id key derivation** — Memory-hard protection against brute-force attacks
- **Secure memory handling** — Sensitive data protected in memory via memguard
- **Session management** — Unlock your vault once, work for 15 minutes without re-entering your master password
- **Password generation** — Built-in cryptographically secure random password generator
- **Clipboard integration** — Passwords copied to clipboard and auto-cleared after 15 seconds
- **Atomic file operations** — Writes go to temp files first, preventing corruption on crash
- **Cross-platform** — Works on Linux, macOS, and Windows

## Security

MSK was designed with security as the top priority. Here is how your data is protected:

### Encryption

Each secret is encrypted individually using **AES-256-GCM** (Galois/Counter Mode), providing both confidentiality and integrity. The encryption key is derived from your master password using **Argon2id** with the following parameters:

| Parameter | Value |
|-----------|-------|
| Time cost | 6 iterations |
| Memory cost | 128 MB |
| Parallelism | 4 threads |
| Key length | 32 bytes (256 bits) |
| Salt length | 16 bytes |

### File Format

Each `.msk` file contains:

```
[MSK magic bytes (3B)] [Version (1B)] [Salt (16B)] [Nonce (12B)] [Encrypted data]
```

All vault files are stored with `0600` permissions (owner read/write only).

### Memory Protection

Sensitive data (master password, decrypted secrets) is held in memory using [memguard](https://github.com/awnumar/memguard), which:

- Prevents memory from being paged to disk
- Encrypts sensitive buffers in RAM
- Automatically purges data on process exit

### Session Security

Sessions expire after **15 minutes**. The session token (32 random bytes) is used to derive an AES key via SHA-256, which encrypts the master key on disk for the session duration.

## Installation

### Download a pre-built binary

Download the latest release for your platform from the [Releases page](https://github.com/amauribechtoldjr/msk/releases).

| Platform | File |
|----------|------|
| Linux (x64) | `msk-<version>-linux-amd64.tar.gz` |
| Linux (ARM64) | `msk-<version>-linux-arm64.tar.gz` |
| macOS (Intel) | `msk-<version>-darwin-amd64.tar.gz` |
| macOS (Apple Silicon) | `msk-<version>-darwin-arm64.tar.gz` |
| Windows (x64) | `msk-<version>-windows-amd64.zip` |

Extract the binary and add it to your `PATH`.

### Build from source

Requires [Go 1.25+](https://go.dev/dl/).

```bash
git clone https://github.com/amauribechtoldjr/msk.git
cd msk
make install
```

This installs `msk` to your `$GOPATH/bin` directory.

## Quick Start

Initialize MSK by setting your master password and vault path:

```bash
msk config
```

You will be prompted to choose a vault path (default: `~/.msk/vault`) and set your master password. The configuration is encrypted and stored in your system's config directory.

Add your first password:

```bash
msk add github
```

Retrieve it (copies to clipboard, auto-clears after 15 seconds):

```bash
msk get github
```

Generate a random password instead of typing one:

```bash
msk add gitlab --generate --length 24
```

Unlock the vault for session-based access (avoids re-entering master password for 15 minutes):

```bash
export MSK_SESSION=$(msk unlock)
```

For a full list of commands and flags, run `msk --help` or `msk <command> --help`.

## How It Works

```
┌─────────────┐    ┌───────────────┐    ┌──────────────┐
│ Master       │───>│ Argon2id      │───>│ 256-bit      │
│ Password     │    │ Key Derivation│    │ Secret Key   │
└─────────────┘    └───────────────┘    └──────┬───────┘
                                               │
┌─────────────┐    ┌───────────────┐           │
│ Plaintext   │───>│ AES-256-GCM   │<──────────┘
│ Secret      │    │ Encrypt       │
└─────────────┘    └───────┬───────┘
                           │
                    ┌──────▼───────┐
                    │ .msk File    │
                    │ (vault dir)  │
                    └──────────────┘
```

Each secret is stored as an individual `.msk` file in your vault directory. File writes are atomic (write to temp file, then rename) to prevent corruption.

**Sessions:** Running `msk unlock` generates a random session token, encrypts your master key with it, and stores the encrypted key on disk. The token is exported as `MSK_SESSION`. Subsequent commands use the session token to decrypt the master key without prompting. Sessions expire after 15 minutes.

## Project Status

MSK is under **active development**. Core functionality (encryption, storage, session management) is stable, but the API and features may evolve. Contributions and feedback are welcome.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/my-feature`)
3. Run tests (`make test`)
4. Commit your changes
5. Open a pull request against `main`

## License

[MIT](LICENSE)
