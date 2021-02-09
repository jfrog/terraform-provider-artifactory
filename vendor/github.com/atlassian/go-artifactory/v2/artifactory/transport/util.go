package transport

import (
	"crypto/sha1"
	"net/http"
)

func getSha1(buf []byte) []byte {
	h := sha1.New()
	h.Write(buf)
	return h.Sum(nil)
}

func deepCopyRequest(req *http.Request, req2 *http.Request) {
	*req2 = *req
	req2.Header = make(http.Header, len(req.Header))
	for k, s := range req.Header {
		req2.Header[k] = append([]string(nil), s...)
	}
}
