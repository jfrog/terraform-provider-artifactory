package retry

import (
	"github.com/go-resty/resty/v2"
	"net/http"
	"regexp"
)

var mergeAndSaveRegex = regexp.MustCompile(".*Could not merge and save new descriptor.*")

func OnMergeError(response *resty.Response, _ error) bool {
	return mergeAndSaveRegex.MatchString(string(response.Body()[:]))
}
func Never(_ *resty.Response, _ error) bool {
	return false
}

func On400Error(response *resty.Response, _ error) bool {
	return response.StatusCode() == 400
}
func On404NotFound(response *resty.Response, _ error) bool {
	return response.StatusCode() == http.StatusNotFound
}
