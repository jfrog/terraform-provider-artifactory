package security

import (
	"github.com/go-resty/resty/v2"
)

func VerifyKeyPair(id string, request *resty.Request) (*resty.Response, error) {
	return request.Head(KeypairEndPoint + id)
}
