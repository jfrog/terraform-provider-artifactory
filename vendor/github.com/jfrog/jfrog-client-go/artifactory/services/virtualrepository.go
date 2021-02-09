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
)

type VirtualRepositoryService struct {
	isUpdate   bool
	client     *rthttpclient.ArtifactoryHttpClient
	ArtDetails auth.ServiceDetails
}

func NewVirtualRepositoryService(client *rthttpclient.ArtifactoryHttpClient, isUpdate bool) *VirtualRepositoryService {
	return &VirtualRepositoryService{client: client, isUpdate: isUpdate}
}

func (vrs *VirtualRepositoryService) GetJfrogHttpClient() *rthttpclient.ArtifactoryHttpClient {
	return vrs.client
}

func (vrs *VirtualRepositoryService) performRequest(params interface{}, repoKey string) error {
	content, err := json.Marshal(params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	httpClientsDetails := vrs.ArtDetails.CreateHttpClientDetails()
	utils.SetContentType("application/vnd.org.jfrog.artifactory.repositories.VirtualRepositoryConfiguration+json", &httpClientsDetails.Headers)
	var url = vrs.ArtDetails.GetUrl() + "api/repositories/" + repoKey
	var operationString string
	var resp *http.Response
	var body []byte
	if vrs.isUpdate {
		log.Info("Updating virtual repository...")
		operationString = "updating"
		resp, body, err = vrs.client.SendPost(url, content, &httpClientsDetails)
	} else {
		log.Info("Creating virtual repository...")
		operationString = "creating"
		resp, body, err = vrs.client.SendPut(url, content, &httpClientsDetails)
	}
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "\n" + clientutils.IndentJson(body)))
	}

	log.Debug("Artifactory response:", resp.Status)
	log.Info("Done " + operationString + " repository.")
	return nil
}

func (vrs *VirtualRepositoryService) Maven(params MavenVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Gradle(params GradleVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Ivy(params IvyVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Sbt(params SbtVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Helm(params HelmVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Rpm(params RpmVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Nuget(params NugetVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Cran(params CranVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Gems(params GemsVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Npm(params NpmVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Bower(params BowerVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Debian(params DebianVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Pypi(params PypiVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Docker(params DockerVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Yum(params YumVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Go(params GoVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) P2(params P2VirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Chef(params ChefVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Puppet(params PuppetVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Conda(params CondaVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Conan(params ConanVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Gitlfs(params GitlfsVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

func (vrs *VirtualRepositoryService) Generic(params GenericVirtualRepositoryParams) error {
	return vrs.performRequest(params, params.Key)
}

type VirtualRepositoryBaseParams struct {
	Key                                           string   `json:"key,omitempty"`
	Rclass                                        string   `json:"rclass"`
	PackageType                                   string   `json:"packageType,omitempty"`
	Description                                   string   `json:"description,omitempty"`
	Notes                                         string   `json:"notes,omitempty"`
	IncludesPattern                               string   `json:"includesPattern,omitempty"`
	ExcludesPattern                               string   `json:"excludesPattern,omitempty"`
	RepoLayoutRef                                 string   `json:"repoLayoutRef,omitempty"`
	Repositories                                  []string `json:"repositories,omitempty"`
	ArtifactoryRequestsCanRetrieveRemoteArtifacts *bool    `json:"artifactoryRequestsCanRetrieveRemoteArtifacts,omitempty"`
	DefaultDeploymentRepo                         string   `json:"defaultDeploymentRepo,omitempty"`
}

type CommonMavenGradleVirtualRepositoryParams struct {
	ForceMavenAuthentication             *bool  `json:"forceMavenAuthentication,omitempty"`
	PomRepositoryReferencesCleanupPolicy string `json:"pomRepositoryReferencesCleanupPolicy,omitempty"`
	KeyPair                              string `json:"keyPair,omitempty"`
}

type MavenVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
	CommonMavenGradleVirtualRepositoryParams
}

func NewMavenVirtualRepositoryParams() MavenVirtualRepositoryParams {
	return MavenVirtualRepositoryParams{VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "maven"}}
}

type GradleVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
	CommonMavenGradleVirtualRepositoryParams
}

func NewGradleVirtualRepositoryParams() GradleVirtualRepositoryParams {
	return GradleVirtualRepositoryParams{VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "gradle"}}
}

type NugetVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
	ForceNugetAuthentication *bool `json:"forceNugetAuthentication,omitempty"`
}

func NewNugetVirtualRepositoryParams() NugetVirtualRepositoryParams {
	return NugetVirtualRepositoryParams{VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "nuget"}}
}

type NpmVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
	ExternalDependenciesEnabled     *bool    `json:"externalDependenciesEnabled,omitempty"`
	ExternalDependenciesPatterns    []string `json:"externalDependenciesPatterns,omitempty"`
	ExternalDependenciesRemoteRepo  string   `json:"externalDependenciesRemoteRepo,omitempty"`
	VirtualRetrievalCachePeriodSecs int      `json:"virtualRetrievalCachePeriodSecs,omitempty"`
}

func NewNpmVirtualRepositoryParams() NpmVirtualRepositoryParams {
	return NpmVirtualRepositoryParams{VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "npm"}}
}

type BowerVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
	ExternalDependenciesEnabled    *bool    `json:"externalDependenciesEnabled,omitempty"`
	ExternalDependenciesPatterns   []string `json:"externalDependenciesPatterns,omitempty"`
	ExternalDependenciesRemoteRepo string   `json:"externalDependenciesRemoteRepo,omitempty"`
}

func NewBowerVirtualRepositoryParams() BowerVirtualRepositoryParams {
	return BowerVirtualRepositoryParams{VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "bower"}}
}

type DebianVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
	DebianTrivialLayout *bool `json:"debianTrivialLayout,omitempty"`
}

func NewDebianVirtualRepositoryParams() DebianVirtualRepositoryParams {
	return DebianVirtualRepositoryParams{VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "debian"}}
}

type GoVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
	ExternalDependenciesEnabled  *bool    `json:"externalDependenciesEnabled,omitempty"`
	ExternalDependenciesPatterns []string `json:"externalDependenciesPatterns,omitempty"`
}

func NewGoVirtualRepositoryParams() GoVirtualRepositoryParams {
	return GoVirtualRepositoryParams{VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "go"}}
}

type ConanVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
	VirtualRetrievalCachePeriodSecs int `json:"virtualRetrievalCachePeriodSecs,omitempty"`
}

func NewConanVirtualRepositoryParams() ConanVirtualRepositoryParams {
	return ConanVirtualRepositoryParams{VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "conan"}}
}

type HelmVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
	VirtualRetrievalCachePeriodSecs int `json:"virtualRetrievalCachePeriodSecs,omitempty"`
}

func NewHelmVirtualRepositoryParams() HelmVirtualRepositoryParams {
	return HelmVirtualRepositoryParams{VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "helm"}}
}

type RpmVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
	VirtualRetrievalCachePeriodSecs int `json:"virtualRetrievalCachePeriodSecs,omitempty"`
}

func NewRpmVirtualRepositoryParams() RpmVirtualRepositoryParams {
	return RpmVirtualRepositoryParams{VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "rpm"}}
}

type CranVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
	VirtualRetrievalCachePeriodSecs int `json:"virtualRetrievalCachePeriodSecs,omitempty"`
}

func NewCranVirtualRepositoryParams() CranVirtualRepositoryParams {
	return CranVirtualRepositoryParams{VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "cran"}}
}

type ChefVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
	VirtualRetrievalCachePeriodSecs int `json:"virtualRetrievalCachePeriodSecs,omitempty"`
}

func NewChefVirtualRepositoryParams() ChefVirtualRepositoryParams {
	return ChefVirtualRepositoryParams{VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "chef"}}
}

type CondaVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
}

func NewCondaVirtualRepositoryParams() CondaVirtualRepositoryParams {
	return CondaVirtualRepositoryParams{VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "conda"}}
}

type GitlfsVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
}

func NewGitlfsVirtualRepositoryParams() GitlfsVirtualRepositoryParams {
	return GitlfsVirtualRepositoryParams{VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "gitlfs"}}
}

type P2VirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
}

func NewP2VirtualRepositoryParams() P2VirtualRepositoryParams {
	return P2VirtualRepositoryParams{VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "p2"}}
}

type GemsVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
}

func NewGemsVirtualRepositoryParams() GemsVirtualRepositoryParams {
	return GemsVirtualRepositoryParams{VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "gems"}}
}

type PypiVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
}

func NewPypiVirtualRepositoryParams() PypiVirtualRepositoryParams {
	return PypiVirtualRepositoryParams{VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "pypi"}}
}

type PuppetVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
}

func NewPuppetVirtualRepositoryParams() PuppetVirtualRepositoryParams {
	return PuppetVirtualRepositoryParams{VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "puppet"}}
}

type IvyVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
}

func NewIvyVirtualRepositoryParams() IvyVirtualRepositoryParams {
	return IvyVirtualRepositoryParams{VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "ivy"}}
}

type SbtVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
}

func NewSbtVirtualRepositoryParams() SbtVirtualRepositoryParams {
	return SbtVirtualRepositoryParams{VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "sbt"}}
}

type DockerVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
}

func NewDockerVirtualRepositoryParams() DockerVirtualRepositoryParams {
	return DockerVirtualRepositoryParams{VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "docker"}}
}

type YumVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
}

func NewYumVirtualRepositoryParams() YumVirtualRepositoryParams {
	return YumVirtualRepositoryParams{VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "yum"}}
}

type GenericVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
}

func NewGenericVirtualRepositoryParams() GenericVirtualRepositoryParams {
	return GenericVirtualRepositoryParams{VirtualRepositoryBaseParams{Rclass: "virtual", PackageType: "generic"}}
}
