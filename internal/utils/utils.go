package utils

import (
	"crypto/rand"
	"math/big"
)

var GenerateCodeEmail = func() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const length = 5

	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		result[i] = letters[n.Int64()]
	}
	return string(result)
}
