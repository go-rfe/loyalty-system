package server

import (
	"crypto/rand"
	"encoding/base32"

	"github.com/go-rfe/logging/log"
)

const (
	tokenSize = 64
)

func getRandomToken() []byte {
	randomBytes := make([]byte, tokenSize)
	token := make([]byte, base32.StdEncoding.EncodedLen(len(randomBytes)))
	if _, err := rand.Read(randomBytes); err != nil {
		log.Fatal().Err(err).Msg("Couldn't get token for JWTAuth")
	}

	base32.StdEncoding.Encode(token, randomBytes)

	return token
}
