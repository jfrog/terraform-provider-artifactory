package transport

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// BasicAuth allows the construction of a HTTP Client that authenticates with basic auth
// It also adds the correct headers to the request
type BasicAuth struct {
	Username  string
	Password  string
	Transport http.RoundTripper
}

// Client returns a HTTP Client and injects the basic auth transport
func (t *BasicAuth) Client() *http.Client {
	return &http.Client{Transport: t}
}

func (t *BasicAuth) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

// RoundTrip allows us to add headers to every request
func (t *BasicAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	// To set extra headers, we must make a copy of the Request so
	// that we don't modify the Request we were given. This is required by the
	// specification of service.RoundTripper.
	//
	// Since we are going to modify only req.Header here, we only need a deep copy
	// of req.Header.
	req2 := new(http.Request)
	deepCopyRequest(req, req2)

	req2.SetBasicAuth(t.Username, t.Password)
	req2.Header.Add(HeaderResultDetail, "info, properties")

	if req.Body != nil {
		reader, _ := req.GetBody()
		buf, _ := ioutil.ReadAll(reader)
		chkSum := getSha1(buf)
		req.Header.Add(HeaderChecksumSha1, fmt.Sprintf("%x", chkSum))
	}

	return t.transport().RoundTrip(req2)
}
