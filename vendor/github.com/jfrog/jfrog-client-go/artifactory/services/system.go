package services

import (
	"encoding/json"
	"errors"
	rthttpclient "github.com/jfrog/jfrog-client-go/artifactory/httpclient"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"net/http"
	"strings"
)

type SystemService struct {
	client     *rthttpclient.ArtifactoryHttpClient
	ArtDetails auth.ServiceDetails
}

func NewSystemService(client *rthttpclient.ArtifactoryHttpClient) *SystemService {
	return &SystemService{client: client}
}

func (ss *SystemService) GetArtifactoryDetails() auth.ServiceDetails {
	return ss.ArtDetails
}

func (ss *SystemService) SetArtifactoryDetails(rt auth.ServiceDetails) {
	ss.ArtDetails = rt
}

func (ss *SystemService) GetJfrogHttpClient() (*rthttpclient.ArtifactoryHttpClient, error) {
	return ss.client, nil
}

func (ss *SystemService) IsDryRun() bool {
	return false
}

func (ss *SystemService) GetVersion() (string, error) {
	httpDetails := ss.ArtDetails.CreateHttpClientDetails()
	resp, body, _, err := ss.client.SendGet(ss.ArtDetails.GetUrl()+"api/system/version", true, &httpDetails)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + utils.IndentJson(body)))
	}
	var version artifactoryVersion
	err = json.Unmarshal(body, &version)
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	return strings.TrimSpace(version.Version), nil
}

func (ss *SystemService) GetServiceId() (string, error) {
	httpDetails := ss.ArtDetails.CreateHttpClientDetails()
	resp, body, _, err := ss.client.SendGet(ss.ArtDetails.GetUrl()+"api/system/service_id", true, &httpDetails)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + utils.IndentJson(body)))
	}

	return string(body), nil
}

type artifactoryVersion struct {
	Version string `json:"version,omitempty"`
}
