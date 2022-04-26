package test

import (
	"math/rand"
	"time"

	"github.com/go-resty/resty/v2"
)

func RandomInt() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(10000000)
}

func RandBool() bool {
	return RandomInt()%2 == 0
}

func RandSelect(items ...interface{}) interface{} {
	return items[RandomInt()%len(items)]
}

func NeverRetry(response *resty.Response, err error) bool {
	return false
}
