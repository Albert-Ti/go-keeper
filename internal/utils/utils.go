package utils

import (
	"crypto/rand"
	"math/big"
)

var GenerateCodeEmail = func() ([]byte, error) {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const length = 5

	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		result[i] = letters[n.Int64()]

		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
