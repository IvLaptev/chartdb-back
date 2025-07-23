package utils

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"
)

func GenerateID(length int64) (string, error) {
	encoding := base32.StdEncoding.WithPadding(base32.NoPadding)

	randBytes := make([]byte, length)
	_, err := rand.Read(randBytes)
	if err != nil {
		return "", fmt.Errorf("read rand bytes: %w", err)
	}

	suffix := encoding.EncodeToString(randBytes)[:length]
	return strings.ToLower(suffix), nil
}
