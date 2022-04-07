package server

import (
	"math/rand"
)

const (
	tokenSize = 48
)

func getRandomToken() []byte {
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890#"

	b := make([]byte, tokenSize)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return b
}
