package utils

import (
	"crypto/subtle"
	"strconv"
	"strings"
)

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

func CheckPass(stored, plain string) bool {
	salt := strings.Split(stored, ".")[0]
	h := HashPass(salt, plain)
	return subtle.ConstantTimeCompare([]byte(h), []byte(stored)) == 1
}
