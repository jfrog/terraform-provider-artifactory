package services

import (
	"encoding/json"
	"errors"
	"net/http"
	"path"

	rthttpclient "github.com/jfrog/jfrog-client-go/artifactory/httpclient"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/auth"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type PromoteService struct {
	client     *rthttpclient.ArtifactoryHttpClient
	ArtDetails auth.ServiceDetails
	DryRun     bool
}

func NewPromotionService(client *rthttpclient.ArtifactoryHttpClient) *PromoteService {
	return &PromoteService{client: client}
}

func (ps *PromoteService) isDryRun() bool {
	return ps.DryRun
}

func (ps *PromoteService) BuildPromote(promotionParams PromotionParams) error {
	message := "Promoting build..."
	if ps.DryRun {
		message = "[Dry run] " + message
	}
	log.Info(message)

	promoteUrl := ps.ArtDetails.GetUrl()
	restApi := path.Join("api/build/promote/", promotionParams.GetBuildName(), promotionParams.GetBuildNumber())
	requestFullUrl, err := utils.BuildArtifactoryUrl(promoteUrl, restApi, make(map[string]string))
	if err != nil {
		return err
	}
	props, err := utils.ParseProperties(promotionParams.GetProperties(), utils.JoinCommas)
	if err != nil {
		return err
	}

	data := BuildPromotionBody{
		Status:              promotionParams.GetStatus(),
		Comment:             promotionParams.GetComment(),
		Copy:                promotionParams.IsCopy(),
		IncludeDependencies: promotionParams.IsIncludeDependencies(),
		SourceRepo:          promotionParams.GetSourceRepo(),
		TargetRepo:          promotionParams.GetTargetRepo(),
		DryRun:              ps.isDryRun(),
		Properties:          props.ToBuildPromoteMap()}
	requestContent, err := json.Marshal(data)
	if err != nil {
		return errorutils.CheckError(err)
	}

	httpClientsDetails := ps.ArtDetails.CreateHttpClientDetails()
	utils.SetContentType("application/vnd.org.jfrog.artifactory.build.PromotionRequest+json", &httpClientsDetails.Headers)

	resp, body, err := ps.client.SendPost(requestFullUrl, requestContent, &httpClientsDetails)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Artifactory response:", resp.Status)
	log.Info("Promoted build", promotionParams.GetBuildName()+"/"+promotionParams.GetBuildNumber(), "to:", promotionParams.GetTargetRepo(), "repository.")
	return nil
}

type BuildPromotionBody struct {
	Comment             string              `json:"comment,omitempty"`
	SourceRepo          string              `json:"sourceRepo,omitempty"`
	TargetRepo          string              `json:"targetRepo,omitempty"`
	Status              string              `json:"status,omitempty"`
	IncludeDependencies bool                `json:"dependencies,omitempty"`
	Copy                bool                `json:"copy,omitempty"`
	DryRun              bool                `json:"dryRun,omitempty"`
	Properties          map[string][]string `json:"properties,omitempty"`
}

type PromotionParams struct {
	BuildName           string
	BuildNumber         string
	TargetRepo          string
	Status              string
	Comment             string
	Copy                bool
	IncludeDependencies bool
	SourceRepo          string
	Properties          string
}

func (bp *PromotionParams) GetBuildName() string {
	return bp.BuildName
}

func (bp *PromotionParams) GetBuildNumber() string {
	return bp.BuildNumber
}

func (bp *PromotionParams) GetTargetRepo() string {
	return bp.TargetRepo
}

func (bp *PromotionParams) GetStatus() string {
	return bp.Status
}

func (bp *PromotionParams) GetComment() string {
	return bp.Comment
}

func (bp *PromotionParams) IsCopy() bool {
	return bp.Copy
}

func (bp *PromotionParams) IsIncludeDependencies() bool {
	return bp.IncludeDependencies
}

func (bp *PromotionParams) GetSourceRepo() string {
	return bp.SourceRepo
}

func (bp *PromotionParams) GetProperties() string {
	return bp.Properties
}

func NewPromotionParams() PromotionParams {
	return PromotionParams{}
}
