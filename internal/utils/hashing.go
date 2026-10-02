package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func RandomHash(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func HashPass(salt string, pass string) string {
	sum := sha256.Sum256([]byte(pass + salt))
	encStr := base64.StdEncoding.EncodeToString(sum[:])
	return fmt.Sprint(salt, ".", encStr)
}
