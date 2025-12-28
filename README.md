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
- Add to your envinronment variables

## Getting Started

First, config your master key, you can use two methods:

Inserting the master key manually:

```
msk -amk "my_manual_designed_masterkey"
```

OR, you can generate a new randomized master key for you (it will be prompted and you can store it anywhere):

```
msk -amk -r
```

To save a new password:

```
msk -p "mynewpassoword" -mk "my-master-key" -n "Name of my password"
```

To get a new password:

```
msk -l "Name of my passoword" -mk "my-master-key"
```
