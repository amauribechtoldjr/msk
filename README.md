# 🔐 MSK – Secure Local Password Manager

**MSK** is a lightweight password manager designed to keep all your credentials **securely stored on your own computer**, without ever exposing them to the internet.

All passwords are **encrypted using a master password**, ensuring that even if someone gains access to your machine, they won’t be able to view any stored data without the correct master key.

To maximize your security:

- Memorize your master password, **or**
- Store it in a safe place **outside your computer**

As long as your master password remains secret, your data stays protected. 🚀

## Installation

- Download
- Build
- Add to your environment variables

## Getting Started

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
