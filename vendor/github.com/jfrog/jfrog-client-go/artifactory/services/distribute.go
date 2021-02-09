package services

import (
	"encoding/json"
	"errors"
	rthttpclient "github.com/jfrog/jfrog-client-go/artifactory/httpclient"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/auth"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"net/http"
	"path"
	"strings"
)

type DistributeService struct {
	client     *rthttpclient.ArtifactoryHttpClient
	ArtDetails auth.ServiceDetails
	DryRun     bool
}

func NewDistributionService(client *rthttpclient.ArtifactoryHttpClient) *DistributeService {
	return &DistributeService{client: client}
}

func (ds *DistributeService) getArtifactoryDetails() auth.ServiceDetails {
	return ds.ArtDetails
}

func (ds *DistributeService) isDryRun() bool {
	return ds.DryRun
}

func (ds *DistributeService) BuildDistribute(params BuildDistributionParams) error {
	dryRun := ""
	if ds.DryRun == true {
		dryRun = "[Dry run] "
	}
	message := "Distributing build..."
	log.Info(dryRun + message)

	distributeUrl := ds.ArtDetails.GetUrl()
	restApi := path.Join("api/build/distribute/", params.GetBuildName(), params.GetBuildNumber())
	requestFullUrl, err := utils.BuildArtifactoryUrl(distributeUrl, restApi, make(map[string]string))
	if err != nil {
		return err
	}

	var sourceRepos []string
	if params.GetSourceRepos() != "" {
		sourceRepos = strings.Split(params.GetSourceRepos(), ",")
	}

	data := BuildDistributionBody{
		SourceRepos:           sourceRepos,
		TargetRepo:            params.GetTargetRepo(),
		Publish:               params.IsPublish(),
		OverrideExistingFiles: params.IsOverrideExistingFiles(),
		GpgPassphrase:         params.GetGpgPassphrase(),
		Async:                 params.IsAsync(),
		DryRun:                ds.isDryRun()}
	requestContent, err := json.Marshal(data)
	if err != nil {
		return errorutils.CheckError(err)
	}

	httpClientsDetails := ds.getArtifactoryDetails().CreateHttpClientDetails()
	utils.SetContentType("application/json", &httpClientsDetails.Headers)

	resp, body, err := ds.client.SendPost(requestFullUrl, requestContent, &httpClientsDetails)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Artifactory response:", resp.Status)
	if params.IsAsync() && !ds.isDryRun() {
		log.Info("Asynchronously distributed build", params.GetBuildName()+"/"+params.GetBuildNumber(), "to:", params.GetTargetRepo(), "repository, logs are available in Artifactory.")
		return nil
	}

	log.Info(dryRun+"Distributed build", params.GetBuildName()+"/"+params.GetBuildNumber(), "to:", params.GetTargetRepo(), "repository.")
	return nil
}

type BuildDistributionParams struct {
	SourceRepos           string
	TargetRepo            string
	GpgPassphrase         string
	Publish               bool
	OverrideExistingFiles bool
	Async                 bool
	BuildName             string
	BuildNumber           string
}

func (bd *BuildDistributionParams) GetSourceRepos() string {
	return bd.SourceRepos
}

func (bd *BuildDistributionParams) GetTargetRepo() string {
	return bd.TargetRepo
}

func (bd *BuildDistributionParams) GetGpgPassphrase() string {
	return bd.GpgPassphrase
}

func (bd *BuildDistributionParams) IsAsync() bool {
	return bd.Async
}

func (bd *BuildDistributionParams) IsPublish() bool {
	return bd.Publish
}

func (bd *BuildDistributionParams) IsOverrideExistingFiles() bool {
	return bd.OverrideExistingFiles
}

func (bd *BuildDistributionParams) GetBuildName() string {
	return bd.BuildName
}

func (bd *BuildDistributionParams) GetBuildNumber() string {
	return bd.BuildNumber
}

type BuildDistributionBody struct {
	SourceRepos           []string `json:"sourceRepos,omitempty"`
	TargetRepo            string   `json:"targetRepo,omitempty"`
	GpgPassphrase         string   `json:"gpgPassphrase,omitempty"`
	Publish               bool     `json:"publish"`
	OverrideExistingFiles bool     `json:"overrideExistingFiles,omitempty"`
	Async                 bool     `json:"async,omitempty"`
	DryRun                bool     `json:"dryRun,omitempty"`
}

func NewBuildDistributionParams() BuildDistributionParams {
	return BuildDistributionParams{}
}
