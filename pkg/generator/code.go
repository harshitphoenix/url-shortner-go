// Package generator provides cryptographically secure short-code generation.
package generator

import (
	"crypto/rand"
	"math/big"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Code returns a random base-62 string of the given length.
// Uses crypto/rand so the output is unpredictable.
func Code(length int) (string, error) {
	max := big.NewInt(int64(len(alphabet)))
	buf := make([]byte, length)
	for i := range buf {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		buf[i] = alphabet[n.Int64()]
	}
	return string(buf), nil
}
