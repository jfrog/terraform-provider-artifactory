package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"

	"github.com/google/go-querystring/query"
)

const (
	userAgent = "go-artifactory"

	MediaTypePlain = "text/plain"
	MediaTypeXml   = "application/xml"

	MediaTypeJson = "application/json"
	MediaTypeForm = "application/x-www-form-urlencoded"
)

type Client struct {
	// HTTP Client used to communicate with the API.
	client *http.Client

	// Base URL for API requests. BaseURL should always be specified with a trailing slash.
	BaseURL *url.URL

	// User agent used when communicating with the Artifactory API.
	UserAgent string
}

// NewClient creates a Client from a provided base url for an artifactory instance and a service Client
func NewClient(baseURL string, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseEndpoint, err := url.Parse(baseURL)

	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(baseEndpoint.Path, "/") {
		baseEndpoint.Path += "/"
	}

	c := &Client{client: httpClient, BaseURL: baseEndpoint, UserAgent: userAgent}
	return c, nil
}

// NewRequest creates an API request. A relative URL can be provided in urlStr, in which case it is resolved relative to the BaseURL
// of the Client. Relative URLs should always be specified without a preceding slash. If specified, the value pointed to
// by body is included as the request body.
func (c *Client) NewRequest(method, urlStr string, body io.Reader) (*http.Request, error) {
	u, err := c.BaseURL.Parse(path.Join(c.BaseURL.Path, urlStr))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	return req, nil
}

func EncodeJson(v interface{}) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	if v == nil {
		return nil, nil
	}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func EncodeURL(body interface{}) (*strings.Reader, error) {
	if body == nil {
		return nil, nil
	}

	urlVals, err := query.Values(body)
	if err != nil {
		return nil, err
	}
	return strings.NewReader(urlVals.Encode()), nil
}

// NewJSONEncodedRequest is a wrapper around Client.NewRequest which encodes the body as a JSON object
func (c *Client) NewJSONEncodedRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	buf, err := EncodeJson(body)
	if err != nil {
		return nil, err
	}
	req, err := c.NewRequest(method, urlStr, buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", MediaTypeJson)
	}
	return req, nil
}

// NewURLEncodedRequest is a wrapper around Client.NewRequest which encodes the body with URL encoding
func (c *Client) NewURLEncodedRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	buf, err := EncodeURL(body)
	if err != nil {
		return nil, err
	}
	req, err := c.NewRequest(method, urlStr, buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", MediaTypeForm)
	}
	return req, nil
}

// Do executes a give request with the given context. If the parameter v is a writer the body will be written to it in
// raw format, else v is assumed to be a struct to unmarshal the body into assuming JSON format. If v is nil then the
// body is not read and can be manually parsed from the response
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	req = req.WithContext(ctx)
	resp, err := c.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if e, ok := err.(*url.Error); ok {
			if url2, err := url.Parse(e.URL); err == nil {
				e.URL = url2.String()
				return nil, e
			}
		}

		return nil, err
	}
	defer resp.Body.Close()

	err = checkResponse(resp)
	if err != nil {
		// even though there was an error, we still return the response
		// in case the caller wants to inspect it further
		return resp, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
			if err == io.EOF {
				err = nil // ignore EOF errors caused by empty response body
			}
		}
	}

	return resp, err
}

func AddOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

// CheckResponse checks the API response for errors, and returns them if present. A response is considered an error if
// it has a status code outside the 200 range. If parsing the response leads to an empty error object, the response will
// be returned as plain text
func checkResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		err = json.Unmarshal(data, errorResponse)
		if err != nil || len(errorResponse.Errors) == 0 {
			return fmt.Errorf(string(data))
		}
	}

	return errorResponse
}
