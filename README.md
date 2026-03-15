# MSK

**A lightweight, offline password manager that keeps your credentials encrypted and local.**

![Go Version](https://img.shields.io/github/go-mod/go-version/amauribechtoldjr/msk)
![License](https://img.shields.io/github/license/amauribechtoldjr/msk)
![CI](https://img.shields.io/github/actions/workflow/status/amauribechtoldjr/msk/ci.yml?label=CI)

MSK stores all your passwords locally on your machine, encrypted with a master password. No cloud, no sync, no network requests — your credentials never leave your computer.

## How It Works

Each secret is stored as an individual encrypted file in your vault directory. Your master password is derived into an encryption key using Argon2id, and each secret is encrypted with AES-256-GCM. Sensitive data in memory is protected using [memguard](https://github.com/awnumar/memguard) to prevent leaks. File writes are atomic to avoid corruption. You can unlock the vault for session-based access, where sessions are time-limited and the master key is encrypted on disk for the session duration.

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

Retrieve it (prints to stdout by default, or use `-c` to copy to clipboard):

```bash
msk get github
msk get github -c
```

Generate a random password instead of typing one:

```bash
msk add gitlab --generate --length 24
```

Unlock the vault (starts a background agent that caches credentials for 15 minutes):

```bash
msk unlock
```

For a full list of commands and flags, run `msk --help` or `msk <command> --help`.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/my-feature`)
3. Run tests (`make test`)
4. Commit your changes
5. Open a pull request against `main`

## License

[MIT](LICENSE)
