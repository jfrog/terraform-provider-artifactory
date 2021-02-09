package _go

import (
	"errors"
	"fmt"

	rthttpclient "github.com/jfrog/jfrog-client-go/artifactory/httpclient"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

type GoService struct {
	client     *rthttpclient.ArtifactoryHttpClient
	ArtDetails auth.ServiceDetails
}

func NewGoService(client *rthttpclient.ArtifactoryHttpClient) *GoService {
	return &GoService{client: client}
}

func (gs *GoService) GetJfrogHttpClient() *rthttpclient.ArtifactoryHttpClient {
	return gs.client
}

func (gs *GoService) SetServiceDetails(artDetails auth.ServiceDetails) {
	gs.ArtDetails = artDetails
}

func (gs *GoService) PublishPackage(params GoParams) error {
	artifactoryVersion, err := gs.ArtDetails.GetVersion()
	if err != nil {
		return err
	}
	publisher := GetCompatiblePublisher(artifactoryVersion)
	if publisher == nil {
		return errorutils.CheckError(errors.New(fmt.Sprintf("Unsupported version of Artifactory: %s", artifactoryVersion)))
	}

	return publisher.PublishPackage(params, gs.client, gs.ArtDetails)
}

type GoParams struct {
	ZipPath    string
	ModPath    string
	InfoPath   string
	ModContent []byte
	Version    string
	Props      string
	TargetRepo string
	ModuleId   string
}

func (gp *GoParams) GetZipPath() string {
	return gp.ZipPath
}

func (gp *GoParams) GetModContent() []byte {
	return gp.ModContent
}

func (gp *GoParams) GetVersion() string {
	return gp.Version
}

func (gp *GoParams) GetProps() string {
	return gp.Props
}

func (gp *GoParams) GetTargetRepo() string {
	return gp.TargetRepo
}

func (gp *GoParams) GetModuleId() string {
	return gp.ModuleId
}

func (gp *GoParams) GetModPath() string {
	return gp.ModPath
}

func (gp *GoParams) GetInfoPath() string {
	return gp.InfoPath
}

func NewGoParams() GoParams {
	return GoParams{}
}
