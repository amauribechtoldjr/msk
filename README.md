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
msk add github -p "my-password"
```

To get the password:

```
msk get github
```

To delete passwords:

```
msk del github
```

TODO:

1 - Add config command to store a file with a single data, that i need to always be able to decrypt first with the master key, and then I can do anything else
.Add this rule for add/get/del/upd

2 - Add second master key prompt confirmation to update/delete commands
