package services

import (
	"encoding/json"
	"errors"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"net/http"

	rthttpclient "github.com/jfrog/jfrog-client-go/artifactory/httpclient"
	"github.com/jfrog/jfrog-client-go/auth"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type PermissionTargetService struct {
	client     *rthttpclient.ArtifactoryHttpClient
	ArtDetails auth.ServiceDetails
}

func NewPermissionTargetService(client *rthttpclient.ArtifactoryHttpClient) *PermissionTargetService {
	return &PermissionTargetService{client: client}
}

func (pts *PermissionTargetService) GetJfrogHttpClient() *rthttpclient.ArtifactoryHttpClient {
	return pts.client
}

func (pts *PermissionTargetService) Delete(permissionTargetName string) error {
	httpClientsDetails := pts.ArtDetails.CreateHttpClientDetails()
	log.Info("Deleting permission target...")
	resp, body, err := pts.client.SendDelete(pts.ArtDetails.GetUrl()+"api/security/permissions/"+permissionTargetName, nil, &httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Artifactory response:", resp.Status)
	log.Info("Done deleting permission target.")
	return nil
}

func (pts *PermissionTargetService) Create(params PermissionTargetParams) error {
	return pts.performRequest(params, false)
}

func (pts *PermissionTargetService) Update(params PermissionTargetParams) error {
	return pts.performRequest(params, true)
}

func (pts *PermissionTargetService) performRequest(params PermissionTargetParams, update bool) error {
	content, err := json.Marshal(params)
	if err != nil {
		return errorutils.CheckError(err)
	}
	httpClientsDetails := pts.ArtDetails.CreateHttpClientDetails()
	utils.SetContentType("application/json", &httpClientsDetails.Headers)
	var url = pts.ArtDetails.GetUrl() + "api/v2/security/permissions/" + params.Name
	var operationString string
	var resp *http.Response
	var body []byte
	if update {
		log.Info("Updating permission target...")
		operationString = "updating"
		resp, body, err = pts.client.SendPut(url, content, &httpClientsDetails)
	} else {
		log.Info("Creating permission target...")
		operationString = "creating"
		resp, body, err = pts.client.SendPost(url, content, &httpClientsDetails)
	}
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}
	log.Debug("Artifactory response:", resp.Status)
	log.Info("Done " + operationString + " permission target.")
	return nil
}

func NewPermissionTargetParams() PermissionTargetParams {
	return PermissionTargetParams{}
}

type PermissionTargetParams struct {
	Name          string                  `json:"name"`
	Repo          PermissionTargetSection `json:"repo,omitempty"`
	Build         PermissionTargetSection `json:"build,omitempty"`
	ReleaseBundle PermissionTargetSection `json:"releaseBundle,omitempty"`
}

type PermissionTargetSection struct {
	IncludePatterns []string `json:"include-patterns,omitempty"`
	ExcludePatterns []string `json:"exclude-patterns,omitempty"`
	Repositories    []string `json:"repositories"`
	Actions         Actions  `json:"actions,omitempty"`
}

type Actions struct {
	Users  map[string][]string `json:"users,omitempty"`
	Groups map[string][]string `json:"groups,omitempty"`
}
