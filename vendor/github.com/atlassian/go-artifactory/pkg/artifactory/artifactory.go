package artifactory

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/google/go-querystring/query"
	"path"
)

const (
	userAgent = "go-artifactory"

	headerChecksumSha1 = "X-Checksum-Sha1"
	headerResultDetail = "X-Result-Detail"
	headerApiToken     = "X-JFrog-Art-Api"

	mediaTypePlain             = "text/plain"
	mediaTypeJson              = "application/json"
	mediaTypeXml               = "application/xml"
	mediaTypeLocalRepository   = "application/vnd.org.jfrog.artifactory.repositories.LocalRepositoryConfiguration+json"
	mediaTypeRemoteRepository  = "application/vnd.org.jfrog.artifactory.repositories.RemoteRepositoryConfiguration+json"
	mediaTypeVirtualRepository = "application/vnd.org.jfrog.artifactory.repositories.VirtualRepositoryConfiguration+json"
	mediaTypeRepositoryDetails = "application/vnd.org.jfrog.artifactory.repositories.RepositoryDetailsList+json"
	mediaTypeSystemVersion     = "application/vnd.org.jfrog.artifactory.system.Version+json"
	mediaTypeUsers             = "application/vnd.org.jfrog.artifactory.security.Users+json"
	mediaTypeUser              = "application/vnd.org.jfrog.artifactory.security.User+json"
	mediaTypeGroups            = "application/vnd.org.jfrog.artifactory.security.Groups+json"
	mediaTypeGroup             = "application/vnd.org.jfrog.artifactory.security.Group+json"
	mediaTypePermissionTargets = "application/vnd.org.jfrog.artifactory.security.PermissionTargets+json"
	mediaTypePermissionTarget  = "application/vnd.org.jfrog.artifactory.security.PermissionTarget+json"
	mediaTypeItemPermissions   = "application/vnd.org.jfrog.artifactory.storage.ItemPermissions+json"
	mediaTypeForm              = "application/x-www-form-urlencoded"
	mediaTypeReplicationConfig = "application/vnd.org.jfrog.artifactory.replications.ReplicationConfigRequest+json"
)

// Client is the container for all the api methods
type Client struct {
	client *http.Client // HTTP client used to communicate with the API.

	// Base URL for API requests. BaseURL should always be specified with a trailing slash.
	BaseURL *url.URL

	// User agent used when communicating with the Artifactory API.
	UserAgent string

	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// Services used for talking to different parts of the Artifactory API.
	Repositories *RepositoriesService
	Security     *SecurityService
	System       *SystemService
	Artifacts    *ArtifactService
}

type service struct {
	client *Client
}

// NewClient creates a Client from a provided base url for an artifactory instance and a http client
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
	c.common.client = c

	c.Repositories = (*RepositoriesService)(&c.common)
	c.Security = (*SecurityService)(&c.common)
	c.System = (*SystemService)(&c.common)
	c.Artifacts = (*ArtifactService)(&c.common)
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

// NewJSONEncodedRequest is a wrapper around client.NewRequest which encodes the body as a JSON object
func (c *Client) NewJSONEncodedRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := c.NewRequest(method, urlStr, buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", mediaTypeJson)
	}
	return req, nil
}

// NewURLEncodedRequest is a wrapper around client.NewRequest which encodes the body with URL encoding
func (c *Client) NewURLEncodedRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	var buf io.Reader
	if body != nil {
		urlVals, err := query.Values(body)
		if err != nil {
			return nil, err
		}
		buf = strings.NewReader(urlVals.Encode())
	}

	req, err := c.NewRequest(method, urlStr, buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", mediaTypeForm)
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

func addOptions(s string, opt interface{}) (string, error) {
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

// ErrorResponse reports one or more errors caused by an API request.
type ErrorResponse struct {
	Response *http.Response `json:"-"`                // HTTP response that caused this error
	Errors   []Status       `json:"errors,omitempty"` // Individual errors
}

// Status is the individual error provided by the API
type Status struct {
	Status  int    `json:"status"`  // Validation error status code
	Message string `json:"message"` // Message describing the error. Errors with Code == "custom" will always have this set.
}

func (e *Status) Error() string {
	return fmt.Sprintf("%d error caused by %s", e.Status, e.Message)
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %+v", r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Errors)
}

// ClientProvider exposes a Client method and enforces a standard for the custom transoirts
type ClientProvider interface {
	Client() *http.Client
}

// BasicAuthTransport allows the construction of a HTTP client that authenticates with basic auth
// It also adds the correct headers to the request
type BasicAuthTransport struct {
	Username  string
	Password  string
	Transport http.RoundTripper
}

// Client returns a HTTP client and injects the basic auth transport
func (t *BasicAuthTransport) Client() *http.Client {
	return &http.Client{Transport: t}
}

func (t *BasicAuthTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

// RoundTrip allows us to add headers to every request
func (t *BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// To set extra headers, we must make a copy of the Request so
	// that we don't modify the Request we were given. This is required by the
	// specification of http.RoundTripper.
	//
	// Since we are going to modify only req.Header here, we only need a deep copy
	// of req.Header.
	req2 := new(http.Request)
	deepCopyRequest(req, req2)

	req2.SetBasicAuth(t.Username, t.Password)
	req2.Header.Add(headerResultDetail, "info, properties")

	if req.Body != nil {
		reader, _ := req.GetBody()
		buf, _ := ioutil.ReadAll(reader)
		chkSum := getSha1(buf)
		req.Header.Add(headerChecksumSha1, fmt.Sprintf("%x", chkSum))
	}

	return t.transport().RoundTrip(req2)
}

// TokenAuthTransport exposes a HTTP client which uses this transport. It authenticates via an Artifactory API token
// It also adds the correct headers to the request
type TokenAuthTransport struct {
	Token     string
	Transport http.RoundTripper
}

// Client returns a HTTP client and injects the token auth transport
func (t *TokenAuthTransport) Client() *http.Client {
	return &http.Client{Transport: t}
}

func (t *TokenAuthTransport) transport() http.RoundTripper {
	if t.Transport != nil {
		return t.Transport
	}
	return http.DefaultTransport
}

// RoundTrip allows us to add headers to every request
func (t *TokenAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// To set extra headers, we must make a copy of the Request so
	// that we don't modify the Request we were given. This is required by the
	// specification of http.RoundTripper.
	//
	// Since we are going to modify only req.Header here, we only need a deep copy
	// of req.Header.
	req2 := new(http.Request)
	deepCopyRequest(req, req2)

	req2.Header.Set(headerApiToken, t.Token)
	req2.Header.Add(headerResultDetail, "info, properties")

	if req.Body != nil {
		reader, _ := req.GetBody()
		buf, _ := ioutil.ReadAll(reader)
		chkSum := getSha1(buf)
		req.Header.Add(headerChecksumSha1, fmt.Sprintf("%x", chkSum))
	}

	return t.transport().RoundTrip(req2)
}

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

// Bool is a helper routine that allocates a new bool value
// to store v and returns a pointer to it.
func Bool(v bool) *bool { return &v }

// Int is a helper routine that allocates a new int value
// to store v and returns a pointer to it.
func Int(v int) *int { return &v }

// Int64 is a helper routine that allocates a new int64 value
// to store v and returns a pointer to it.
func Int64(v int64) *int64 { return &v }

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string { return &v }
