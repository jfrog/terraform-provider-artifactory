package client

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/go-resty/resty/v2"
)

func Build(URL, version string) (*resty.Client, error) {
	u, err := url.ParseRequestURI(URL)

	if err != nil {
		return nil, err
	}

	baseUrl := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	restyBase := resty.New().SetHostURL(baseUrl).OnAfterResponse(func(client *resty.Client, response *resty.Response) error {
		if response == nil {
			return fmt.Errorf("no response found")
		}

		if response.StatusCode() >= http.StatusBadRequest {
			return fmt.Errorf("\n%d %s %s\n%s", response.StatusCode(), response.Request.Method, response.Request.URL, string(response.Body()[:]))
		}
		return nil
	}).
		SetHeader("content-type", "application/json").
		SetHeader("accept", "*/*").
		SetHeader("user-agent", "jfrog/terraform-provider-artifactory:"+version).
		SetRetryCount(20)

	restyBase.DisableWarn = true

	return restyBase, nil
}

func AddAuth(client *resty.Client, apiKey, accessToken string) (*resty.Client, error) {
	if accessToken != "" {
		return client.SetAuthToken(accessToken), nil
	}
	if apiKey != "" {
		return client.SetHeader("X-JFrog-Art-Api", apiKey), nil
	}
	return nil, fmt.Errorf("no authentication details supplied")
}

var mergeAndSaveRegex = regexp.MustCompile(".*Could not merge and save new descriptor.*")

func RetryOnMergeError(response *resty.Response, _r error) bool {
	return mergeAndSaveRegex.MatchString(string(response.Body()[:]))
}

func NeverRetry(response *resty.Response, err error) bool {
	return false
}
