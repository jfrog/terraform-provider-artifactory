package services

import (
	"errors"
	rthttpclient "github.com/jfrog/jfrog-client-go/artifactory/httpclient"
	"github.com/jfrog/jfrog-client-go/auth"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"net/http"
)

type DeleteRepositoryService struct {
	client     *rthttpclient.ArtifactoryHttpClient
	ArtDetails auth.ServiceDetails
}

func NewDeleteRepositoryService(client *rthttpclient.ArtifactoryHttpClient) *DeleteRepositoryService {
	return &DeleteRepositoryService{client: client}
}

func (drs *DeleteRepositoryService) GetJfrogHttpClient() *rthttpclient.ArtifactoryHttpClient {
	return drs.client
}

func (drs *DeleteRepositoryService) Delete(repoKey string) error {
	httpClientsDetails := drs.ArtDetails.CreateHttpClientDetails()
	log.Info("Deleting repository...")
	resp, body, err := drs.client.SendDelete(drs.ArtDetails.GetUrl()+"api/repositories/"+repoKey, nil, &httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Artifactory response:", resp.Status)
	log.Info("Done deleting repository.")
	return nil
}
