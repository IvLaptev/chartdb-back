package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var key = []byte("1234567890123456")
var message = "Hello, world!"

func TestAES128(t *testing.T) {
	encrypted, err := AES128Encrypt(message, key)
	if err != nil {
		t.Error(err)
	}

	decrypted, err := AES128Decrypt(encrypted, key)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, message, decrypted)
}

func TestSHA1(t *testing.T) {
	assert.Equal(t, "943a702d06f34599aee1f8da8ef9f7296031d699", SHA1(message))
}
