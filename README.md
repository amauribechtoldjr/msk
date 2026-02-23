# MSK – Lightweight Local Password Manager

**MSK** is a lightweight password manager designed to keep all your credentials **securely stored on your own computer**, without ever exposing them to the internet.

All passwords are **encrypted using a master password**, ensuring that even if someone gains access to your machine, they won’t be able to view any stored data without the correct master key.

To maximize your security:

- Memorize your master password, **or**
- Store it in a safe place **outside your computer**

As long as your master password remains secret, your data stays protected.

## Installation

- Download
- Build
- Add to your environment variables

## Getting Started

Before using MSK, you need to initialize it by running the config command. This sets your master password and vault path (where encrypted files are stored):

```
msk config
```

You'll be prompted to choose a vault path (press Enter to use the default `~/.msk/vault`) and set your master password. The config is encrypted and stored in your system's config directory.

You can re-run `msk config` at any time to change your vault path or master password.

## Usage

To save a new password:

```
msk add github
```

To update a password:

```
msk update github
```

To get the password (copies to clipboard):

```
msk get github
```

To delete a password:

```
msk del github
```

To list all stored passwords:

```
msk list
```
