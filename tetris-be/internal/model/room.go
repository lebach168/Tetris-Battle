package model

import (
	"crypto/rand"
	"math/big"
)

type Room struct {
	ID  string   `json:"id"`
	P   []Player `json:"p"`
	Key string   `json:"-"`
}

func GenerateRoomKey(n int) (string, error) {
	b := make([]byte, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		b[i] = letters[num.Int64()]
	}
	return string(b), nil
}

const letters = "abcdefghijklmnopqrstxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
