package auth

import (
	"sync"
	"time"

	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
)

var expiryHandleMutex sync.Mutex

// Implement this function and append it to create an interceptor that will run pre request in the http client
type PreRequestInterceptorFunc func(*CommonConfigFields, *httputils.HttpClientDetails) error

type ServiceDetails interface {
	GetUrl() string
	GetUser() string
	GetPassword() string
	GetApiKey() string
	GetAccessToken() string
	GetPreRequestInterceptor() []PreRequestInterceptorFunc
	GetClientCertPath() string
	GetClientCertKeyPath() string
	GetSshUrl() string
	GetSshKeyPath() string
	GetSshPassphrase() string
	GetSshAuthHeaders() map[string]string
	GetVersion() (string, error)

	SetUrl(url string)
	SetUser(user string)
	SetPassword(password string)
	SetApiKey(apiKey string)
	SetAccessToken(accessToken string)
	AppendPreRequestInterceptor(PreRequestInterceptorFunc)
	SetClientCertPath(certificatePath string)
	SetClientCertKeyPath(certificatePath string)
	SetSshUrl(url string)
	SetSshKeyPath(sshKeyPath string)
	SetSshPassphrase(sshPassphrase string)
	SetSshAuthHeaders(sshAuthHeaders map[string]string)

	IsSshAuthHeaderSet() bool
	IsSshAuthentication() bool
	AuthenticateSsh(sshKey, sshPassphrase string) error
	InitSsh() error
	RunPreRequestInterceptors(httpClientDetails *httputils.HttpClientDetails) error

	CreateHttpClientDetails() httputils.HttpClientDetails
}

type CommonConfigFields struct {
	Url                    string                      `json:"-"`
	User                   string                      `json:"-"`
	Password               string                      `json:"-"`
	ApiKey                 string                      `json:"-"`
	AccessToken            string                      `json:"-"`
	PreRequestInterceptors []PreRequestInterceptorFunc `json:"-"`
	ClientCertPath         string                      `json:"-"`
	ClientCertKeyPath      string                      `json:"-"`
	Version                string                      `json:"-"`
	SshUrl                 string                      `json:"-"`
	SshKeyPath             string                      `json:"-"`
	SshPassphrase          string                      `json:"-"`
	SshAuthHeaders         map[string]string           `json:"-"`
	TokenMutex             sync.Mutex
}

func (ccf *CommonConfigFields) GetUrl() string {
	return ccf.Url
}

func (ccf *CommonConfigFields) GetUser() string {
	return ccf.User
}

func (ccf *CommonConfigFields) GetPassword() string {
	return ccf.Password
}

func (ccf *CommonConfigFields) GetApiKey() string {
	return ccf.ApiKey
}

func (ccf *CommonConfigFields) GetAccessToken() string {
	return ccf.AccessToken
}

func (ccf *CommonConfigFields) GetPreRequestInterceptor() []PreRequestInterceptorFunc {
	return ccf.PreRequestInterceptors
}

func (ccf *CommonConfigFields) GetClientCertPath() string {
	return ccf.ClientCertPath
}

func (ccf *CommonConfigFields) GetClientCertKeyPath() string {
	return ccf.ClientCertKeyPath
}

func (ccf *CommonConfigFields) GetSshUrl() string {
	return ccf.SshUrl
}

func (ccf *CommonConfigFields) GetSshKeyPath() string {
	return ccf.SshKeyPath
}

func (ccf *CommonConfigFields) GetSshPassphrase() string {
	return ccf.SshPassphrase
}

func (ccf *CommonConfigFields) GetSshAuthHeaders() map[string]string {
	return ccf.SshAuthHeaders
}

func (ccf *CommonConfigFields) SetUrl(url string) {
	ccf.Url = url
}

func (ccf *CommonConfigFields) SetUser(user string) {
	ccf.User = user
}

func (ccf *CommonConfigFields) SetPassword(password string) {
	ccf.Password = password
}

func (ccf *CommonConfigFields) SetApiKey(apiKey string) {
	ccf.ApiKey = apiKey
}

func (ccf *CommonConfigFields) SetAccessToken(accessToken string) {
	ccf.AccessToken = accessToken
}

func (ccf *CommonConfigFields) AppendPreRequestInterceptor(interceptor PreRequestInterceptorFunc) {
	ccf.PreRequestInterceptors = append(ccf.PreRequestInterceptors, interceptor)
}

func (ccf *CommonConfigFields) SetClientCertPath(certificatePath string) {
	ccf.ClientCertPath = certificatePath
}

func (ccf *CommonConfigFields) SetClientCertKeyPath(certificatePath string) {
	ccf.ClientCertKeyPath = certificatePath
}

func (ccf *CommonConfigFields) SetSshUrl(sshUrl string) {
	ccf.SshUrl = sshUrl
}

func (ccf *CommonConfigFields) SetSshKeyPath(sshKeyPath string) {
	ccf.SshKeyPath = sshKeyPath
}

func (ccf *CommonConfigFields) SetSshPassphrase(sshPassphrase string) {
	ccf.SshPassphrase = sshPassphrase
}

func (ccf *CommonConfigFields) SetSshAuthHeaders(sshAuthHeaders map[string]string) {
	ccf.SshAuthHeaders = sshAuthHeaders
}

func (ccf *CommonConfigFields) IsSshAuthHeaderSet() bool {
	return len(ccf.SshAuthHeaders) > 0
}

func (ccf *CommonConfigFields) IsSshAuthentication() bool {
	return fileutils.IsSshUrl(ccf.Url) || ccf.SshUrl != ""
}

func (ccf *CommonConfigFields) AuthenticateSsh(sshKeyPath, sshPassphrase string) error {
	// If SshUrl is unset, set it and use it to authenticate.
	// The SshUrl variable could be used again later if there's a need to reauthenticate (Url is being overwritten with baseUrl).
	if ccf.SshUrl == "" {
		ccf.SshUrl = ccf.Url
	}

	sshHeaders, baseUrl, err := SshAuthentication(ccf.SshUrl, sshKeyPath, sshPassphrase)
	if err != nil {
		return err
	}

	// Set base url as the connection url
	ccf.Url = baseUrl
	ccf.SetSshAuthHeaders(sshHeaders)
	return nil
}

func (ccf *CommonConfigFields) InitSsh() error {
	if ccf.IsSshAuthentication() {
		if !ccf.IsSshAuthHeaderSet() {
			err := ccf.AuthenticateSsh(ccf.SshKeyPath, ccf.SshPassphrase)
			if err != nil {
				return err
			}
		}
		ccf.AppendPreRequestInterceptor(SshTokenRefreshPreRequestInterceptor)
	}
	return nil
}

// Runs an interceptor before sending a request via the http client
func (ccf *CommonConfigFields) RunPreRequestInterceptors(httpClientDetails *httputils.HttpClientDetails) error {
	for _, exec := range ccf.PreRequestInterceptors {
		err := exec(ccf, httpClientDetails)
		if err != nil {
			return err
		}
	}
	return nil
}

// Handles the process of acquiring a new ssh token
func SshTokenRefreshPreRequestInterceptor(fields *CommonConfigFields, httpClientDetails *httputils.HttpClientDetails) error {
	if !fields.IsSshAuthentication() {
		return nil
	}
	curToken := httpClientDetails.Headers["Authorization"]
	timeLeft, err := GetTokenMinutesLeft(curToken)
	if err != nil || timeLeft > RefreshBeforeExpiryMinutes {
		return err
	}

	// Lock expiryHandleMutex to make sure only one authentication is made.
	expiryHandleMutex.Lock()
	defer expiryHandleMutex.Unlock()
	// Reauthenticate only if a new token wasn't acquired (by another thread) while waiting at mutex.
	if fields.GetSshAuthHeaders()["Authorization"] == curToken {
		// If token isn't already expired, Wait to make sure requests using the current token are sent before it is refreshed and becomes invalid.
		if timeLeft != 0 {
			time.Sleep(WaitBeforeRefreshSeconds * time.Second)
		}

		// Obtain a new token.
		err := fields.AuthenticateSsh(fields.GetSshKeyPath(), fields.GetSshPassphrase())
		if err != nil {
			return err
		}
	}

	// Copy new token from the mutual headers map in ServiceDetails to the private headers map in httpClientDetails
	utils.MergeMaps(fields.GetSshAuthHeaders(), httpClientDetails.Headers)
	return nil
}

func (ccf *CommonConfigFields) CreateHttpClientDetails() httputils.HttpClientDetails {
	return httputils.HttpClientDetails{
		User:        ccf.User,
		Password:    ccf.Password,
		ApiKey:      ccf.ApiKey,
		AccessToken: ccf.AccessToken,
		Headers:     utils.CopyMap(ccf.GetSshAuthHeaders())}
}
