package services

import (
	"io"

	rthttpclient "github.com/jfrog/jfrog-client-go/artifactory/httpclient"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/auth"
)

type AqlService struct {
	client     *rthttpclient.ArtifactoryHttpClient
	ArtDetails auth.ServiceDetails
}

func NewAqlService(client *rthttpclient.ArtifactoryHttpClient) *AqlService {
	return &AqlService{client: client}
}

func (s *AqlService) GetArtifactoryDetails() auth.ServiceDetails {
	return s.ArtDetails
}

func (s *AqlService) SetArtifactoryDetails(rt auth.ServiceDetails) {
	s.ArtDetails = rt
}

func (s *AqlService) IsDryRun() bool {
	return false
}

func (s *AqlService) GetJfrogHttpClient() (*rthttpclient.ArtifactoryHttpClient, error) {
	return s.client, nil
}

func (s *AqlService) ExecAql(aql string) (io.ReadCloser, error) {
	return s.exec(aql)
}

func (s *AqlService) exec(aql string) (io.ReadCloser, error) {
	return utils.ExecAql(aql, s)
}
