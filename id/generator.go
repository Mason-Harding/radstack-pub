package id

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const (
	idBytesAlphaNumLower = "0123456789abcdefghijklmnopqrstuvwxyz"
	idBytesNum           = "0123456789"
)

func RandStringAlphaNumLower(n int) string {
	return RandString(n, idBytesAlphaNumLower)
}

func RandStringBytesNumOnly(n int) string {
	return RandString(n, idBytesNum)
}

func RandString(n int, fromChars string) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = fromChars[rand.Intn(len(fromChars))]
	}
	return string(b)
}
