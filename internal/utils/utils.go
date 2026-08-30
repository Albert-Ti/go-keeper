package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

var GenerateUUID = func() string {
	key := make([]byte, 9)
	rand.Read(key)
	return base64.RawURLEncoding.EncodeToString(key)
}

func GenerateMockUUID() func() {
	original := GenerateUUID
	counter := 0

	GenerateUUID = func() string {
		counter++
		return "key_" + string(rune('0'+counter))
	}

	return func() {
		GenerateUUID = original
	}
}

func RandomHash(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func HashPassword(salt string, pass string) string {
	sum := sha256.Sum256([]byte(pass + salt))
	encStr := base64.StdEncoding.EncodeToString(sum[:])
	return fmt.Sprint(salt, ".", encStr)
}

func AlgoLuna(order string) bool {
	order = strings.TrimSpace(order)

	var sum int
	var isSecond bool

	for i := len(order) - 1; i >= 0; i-- {
		n, err := strconv.Atoi(string(order[i]))
		if err != nil {
			return false
		}

		if isSecond {
			n *= 2
			if n >= 10 {
				n = n/10 + n%10
			}
		}

		sum += n
		isSecond = !isSecond
	}

	return sum%10 == 0
}
