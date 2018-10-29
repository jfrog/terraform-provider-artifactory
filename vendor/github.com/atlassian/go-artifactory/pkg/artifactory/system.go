package artifactory

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

type SystemService service

// System Info
// Get general system information.
// Since: 2.2.0
// Security: Requires a valid admin user
func (s *SystemService) GetSystemInfo(ctx context.Context) (*string, *http.Response, error) {
	path := "/api/system"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypePlain)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

// Get a simple status response about the state of Artifactory
// Returns 200 code with an 'OK' text if Artifactory is working properly, if not will return an HTTP error code with a reason.
// Since: 2.3.0
// Security: Requires a valid user (can be anonymous).  If artifactory.ping.allowUnauthenticated=true is set in
// 			 artifactory.system.properties, then no authentication is required even if anonymous access is disabled.
func (s *SystemService) Ping(ctx context.Context) (*string, *http.Response, error) {
	path := "/api/system/ping"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypePlain)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

type VerifyConnectionOptions struct {
	Endpoint *string `json:"endpoint,omitempty"` // Mandatory
	Username *string `json:"username,omitempty"` // Optional
	Password *string `json:"password,omitempty"` // Optional
}

// Verifies a two-way connection between Artifactory and another product
// Returns Success (200) if Artifactory receives a similar success code (200) from the provided endpoint.
// Upon error, returns 400 along with a JSON object that contains the error returned from the other system.
// Since: 4.15.0
// Security: Requires an admin user.
func (s *SystemService) VerifyConnection(ctx context.Context, opt *VerifyConnectionOptions) (*string, *http.Response, error) {
	url := "/api/system/verifyconnection"

	req, err := s.client.NewJSONEncodedRequest("POST", url, opt)
	if err != nil {
		return nil, nil, err
	}

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}

	return String(buf.String()), resp, nil
}

// Get the general configuration (artifactory.config.xml).
// Since: 2.2.0
// Security: Requires a valid admin user
func (s *SystemService) GetConfiguration(ctx context.Context) (*string, *http.Response, error) {
	path := "/api/system/configuration"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeXml)

	buf := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, buf)
	if err != nil {
		return nil, resp, err
	}
	return String(buf.String()), resp, nil
}

// Changes the Custom URL base
// Since: 3.9.0
// Security: Requires a valid admin user
func (s *SystemService) UpdateUrlBase(ctx context.Context, newUrl string) (*http.Response, error) {
	path := "/api/system/configuration/baseUrl"
	req, err := s.client.NewJSONEncodedRequest("PUT", path, newUrl)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-type", mediaTypePlain)

	return s.client.Do(ctx, req, nil)
}

type LicenseDetails struct {
	Type         *string `json:"type,omitempty"`
	ValidThrough *string `json:"validThrough,omitempty"`
	LicensedTo   *string `json:"licensedTo,omitempty"`
}

func (r LicenseDetails) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

// Retrieve information about the currently installed license.
// Since: 3.3.0
// Security: Requires a valid admin user
func (s *SystemService) GetLicense(ctx context.Context) (*LicenseDetails, *http.Response, error) {
	path := "/api/system/license"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	v := new(LicenseDetails)
	resp, err := s.client.Do(ctx, req, v)
	return v, resp, err
}

type LicenseKey struct {
	LicenseKey *string `json:"licenseKey,omitempty"`
}

// Install new license key or change the current one.
// Since: 3.3.0
// Security: Requires a valid admin user
func (s *SystemService) InstallLicense(ctx context.Context, licenseKey *LicenseKey) (*Status, *http.Response, error) {
	path := "/api/system/licenses"
	req, err := s.client.NewJSONEncodedRequest("POST", path, licenseKey)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	v := new(Status)
	resp, err := s.client.Do(ctx, req, v)
	return v, resp, err
}

type HALicense struct {
	Type         *string `json:"type,omitempty"`
	ValidThrough *string `json:"validThrough,omitempty"` // validity date formatted MMM DD, YYYY
	LicensedTo   *string `json:"licensedTo,omitempty"`
	LicenseHash  *string `json:"licenseHash,omitempty"`
	NodeId       *string `json:"nodeId,omitempty"`  // Node ID of the node activated with this license | Not in use
	NodeUrl      *string `json:"nodeUrl,omitempty"` // URL of the node activated with this license | Not in use
	Expired      *bool   `json:"expired,omitempty"`
}

type HALicenses struct {
	Licenses *[]HALicense `json:"licenses,omitempty"`
}

func (r HALicenses) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

// Retrieve information about the currently installed licenses in an HA cluster.
// Since: 5.0.0
// Security: Requires a valid admin user
func (s *SystemService) ListHALicenses(ctx context.Context) (*HALicenses, *http.Response, error) {
	path := "/api/system/licenses"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	v := new(HALicenses)
	resp, err := s.client.Do(ctx, req, v)
	return v, resp, err
}

// Install a new license key(s) on an HA cluster.
// Since: 5.0.0
// Security: Requires an admin user
func (s *SystemService) InstallHALicenses(ctx context.Context, licenses []LicenseKey) (*Status, *http.Response, error) {
	path := "/api/system/licenses"
	req, err := s.client.NewJSONEncodedRequest("POST", path, licenses)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	v := new(Status)
	resp, err := s.client.Do(ctx, req, v)
	return v, resp, err
}

type HALicenseHashes struct {
	LicenseHash *[]string `url:"licenseHash,omitempty"`
}

// Deletes a license key from an HA cluster.
// Since: 5.0.0
// Security: Requires an admin user
func (s *SystemService) DeleteHALicenses(ctx context.Context, licenseHashes HALicenseHashes) (*Status, *http.Response, error) {
	path, err := addOptions("/api/system/licenses", licenseHashes)
	if err != nil {
		return nil, nil, err
	}
	req, err := s.client.NewRequest("DELETE", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	v := new(Status)
	resp, err := s.client.Do(ctx, req, v)
	return v, resp, err
}

type VersionAddOns struct {
	Version  *string   `json:"version,omitempty"`
	Revision *string   `json:"revision,omitempty"`
	Addons   *[]string `json:"addons,omitempty"`
}

func (r VersionAddOns) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

// Retrieve information about the current Artifactory version, revision, and currently installed Add-ons
// Since: 2.2.2
// Security: Requires a valid user (can be anonymous)
func (s *SystemService) GetVersionAndAddons(ctx context.Context) (*VersionAddOns, *http.Response, error) {
	path := "/api/system/version"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeSystemVersion)

	v := new(VersionAddOns)
	resp, err := s.client.Do(ctx, req, v)
	return v, resp, err
}

type ReverseProxyConfig struct {
	Key                      *string `json:"key,omitempty"`
	WebServerType            *string `json:"webServerType,omitempty"`
	ArtifactoryAppContext    *string `json:"artifactoryAppContext,omitempty"`
	PublicAppContext         *string `json:"publicAppContext,omitempty"`
	ServerName               *string `json:"serverName,omitempty"`
	ServerNameExpression     *string `json:"serverNameExpression,omitempty"`
	ArtifactoryServerName    *string `json:"artifactoryServerName,omitempty"`
	ArtifactoryPort          *int    `json:"artifactoryPort,omitempty"`
	SslCertificate           *string `json:"sslCertificate,omitempty"`
	SslKey                   *string `json:"sslKey,omitempty"`
	DockerReverseProxyMethod *string `json:"dockerReverseProxyMethod,omitempty"`
	UseHttps                 *bool   `json:"useHttps,omitempty"`
	UseHttp                  *bool   `json:"useHttp,omitempty"`
	SslPort                  *int    `json:"sslPort,omitempty"`
	HttpPort                 *int    `json:"httpPort,omitempty"`
}

func (r ReverseProxyConfig) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

// Retrieves the reverse proxy configuration
// Since: 4.3.1
// Security: Requires a valid admin user
func (s *SystemService) GetReverseProxyConfig(ctx context.Context) (*ReverseProxyConfig, *http.Response, error) {
	path := "/api/system/configuration/webServer"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeJson)

	v := new(ReverseProxyConfig)
	resp, err := s.client.Do(ctx, req, v)
	return v, resp, err
}

// Updates the reverse proxy configuration
// Since: 4.3.1
// Security: Requires an admin user
func (s *SystemService) UpdateReverseProxyConfig(ctx context.Context, config *ReverseProxyConfig) (*http.Response, error) {
	path := "/api/system/configuration/webServer"
	req, err := s.client.NewJSONEncodedRequest("POST", path, config)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Gets the reverse proxy configuration snippet in text format
// Since: 4.3.1
// Security: Requires a valid user (not anonymous)
func (s *SystemService) GetReverseProxySnippet(ctx context.Context) (*string, *http.Response, error) {
	path := "/api/system/configuration/reverseProxy/nginx"
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypePlain)

	v := new(bytes.Buffer)
	resp, err := s.client.Do(ctx, req, v)
	return String(v.String()), resp, err
}
