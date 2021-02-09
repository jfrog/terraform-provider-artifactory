package services

import (
	"io"
	"net/http"

	rthttpclient "github.com/jfrog/jfrog-client-go/artifactory/httpclient"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

type ReadFileService struct {
	client       *rthttpclient.ArtifactoryHttpClient
	ArtDetails   auth.ServiceDetails
	DryRun       bool
	MinSplitSize int64
	SplitCount   int
}

func NewReadFileService(client *rthttpclient.ArtifactoryHttpClient) *ReadFileService {
	return &ReadFileService{client: client}
}

func (ds *ReadFileService) GetArtifactoryDetails() auth.ServiceDetails {
	return ds.ArtDetails
}

func (ds *ReadFileService) SetArtifactoryDetails(rt auth.ServiceDetails) {
	ds.ArtDetails = rt
}

func (ds *ReadFileService) IsDryRun() bool {
	return ds.DryRun
}

func (ds *ReadFileService) GetJfrogHttpClient() (*rthttpclient.ArtifactoryHttpClient, error) {
	return ds.client, nil
}

func (ds *ReadFileService) SetServiceDetails(artDetails auth.ServiceDetails) {
	ds.ArtDetails = artDetails
}

func (ds *ReadFileService) SetDryRun(isDryRun bool) {
	ds.DryRun = isDryRun
}

func (ds *ReadFileService) setMinSplitSize(minSplitSize int64) {
	ds.MinSplitSize = minSplitSize
}

func (ds *ReadFileService) ReadRemoteFile(downloadPath string) (io.ReadCloser, error) {
	readPath, err := utils.BuildArtifactoryUrl(ds.ArtDetails.GetUrl(), downloadPath, make(map[string]string))
	if err != nil {
		return nil, err
	}
	httpClientsDetails := ds.ArtDetails.CreateHttpClientDetails()
	ioReadCloser, resp, err := ds.client.ReadRemoteFile(readPath, &httpClientsDetails)
	if err != nil {
		return nil, err
	}
	err = errorutils.CheckResponseStatus(resp, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return ioReadCloser, err
}
