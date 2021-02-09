package services

import (
	"encoding/json"
	"errors"
	"net/http"

	rthttpclient "github.com/jfrog/jfrog-client-go/artifactory/httpclient"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/auth"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type GetReplicationService struct {
	client     *rthttpclient.ArtifactoryHttpClient
	ArtDetails auth.ServiceDetails
}

func NewGetReplicationService(client *rthttpclient.ArtifactoryHttpClient) *GetReplicationService {
	return &GetReplicationService{client: client}
}

func (drs *GetReplicationService) GetJfrogHttpClient() *rthttpclient.ArtifactoryHttpClient {
	return drs.client
}

func (drs *GetReplicationService) GetReplication(repoKey string) ([]utils.ReplicationParams, error) {
	body, err := drs.preform(repoKey)
	if err != nil {
		return nil, err
	}
	var replicationConf []utils.ReplicationParams
	if err := json.Unmarshal(body, &replicationConf); err != nil {
		return nil, errorutils.CheckError(err)
	}
	return replicationConf, nil
}

func (drs *GetReplicationService) preform(repoKey string) ([]byte, error) {
	httpClientsDetails := drs.ArtDetails.CreateHttpClientDetails()
	log.Info("Retrieve replication configuration...")
	resp, body, _, err := drs.client.SendGet(drs.ArtDetails.GetUrl()+"api/replications/"+repoKey, true, &httpClientsDetails)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}
	log.Debug("Artifactory response:", resp.Status)
	log.Info("Done retrieve replication job.")
	return body, nil
}
