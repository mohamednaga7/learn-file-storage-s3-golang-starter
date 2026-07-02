package main

import (
	"crypto/rand"
	"encoding/base64"
)

func generateRandomKey() string {
	tokenBytes := make([]byte, 32)

	_, _ = rand.Read(tokenBytes)

	return base64.RawURLEncoding.EncodeToString(tokenBytes)
}
