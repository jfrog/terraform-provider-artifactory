package httputils

import (
	"github.com/jfrog/jfrog-client-go/utils"
	"net/http"
)

type HttpClientDetails struct {
	User        string
	Password    string
	ApiKey      string
	AccessToken string
	Headers     map[string]string
	Transport   *http.Transport
}

func (httpClientDetails HttpClientDetails) Clone() *HttpClientDetails {
	headers := make(map[string]string)
	utils.MergeMaps(httpClientDetails.Headers, headers)
	return &HttpClientDetails{
		User:        httpClientDetails.User,
		Password:    httpClientDetails.Password,
		ApiKey:      httpClientDetails.ApiKey,
		AccessToken: httpClientDetails.AccessToken,
		Headers:     headers}
}
