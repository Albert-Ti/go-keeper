package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
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

func ValidatePass(pass string) {

}
