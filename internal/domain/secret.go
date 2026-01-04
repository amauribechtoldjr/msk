package domain

import "time"

type Secret struct {
	Name      string
	Password  []byte
	CreatedAt time.Time
}

type EncryptedSecret struct {
	Data []byte
	Salt [16]byte
	Nonce [12]byte
}