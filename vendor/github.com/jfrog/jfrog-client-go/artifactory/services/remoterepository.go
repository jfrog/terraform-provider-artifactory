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

type RemoteRepositoryService struct {
	isUpdate   bool
	client     *rthttpclient.ArtifactoryHttpClient
	ArtDetails auth.ServiceDetails
}

func NewRemoteRepositoryService(client *rthttpclient.ArtifactoryHttpClient, isUpdate bool) *RemoteRepositoryService {
	return &RemoteRepositoryService{client: client, isUpdate: isUpdate}
}

func (rrs *RemoteRepositoryService) GetJfrogHttpClient() *rthttpclient.ArtifactoryHttpClient {
	return rrs.client
}

func (rrs *RemoteRepositoryService) performRequest(params interface{}, repoKey string) error {
	content, err := json.Marshal(params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	httpClientsDetails := rrs.ArtDetails.CreateHttpClientDetails()
	utils.SetContentType("application/vnd.org.jfrog.artifactory.repositories.RemoteRepositoryConfiguration+json", &httpClientsDetails.Headers)
	var url = rrs.ArtDetails.GetUrl() + "api/repositories/" + repoKey
	var operationString string
	var resp *http.Response
	var body []byte
	if rrs.isUpdate {
		log.Info("Updating remote repository...")
		operationString = "updating"
		resp, body, err = rrs.client.SendPost(url, content, &httpClientsDetails)
	} else {
		log.Info("Creating remote repository...")
		operationString = "creating"
		resp, body, err = rrs.client.SendPut(url, content, &httpClientsDetails)
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

func (rrs *RemoteRepositoryService) Maven(params MavenRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Gradle(params GradleRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Ivy(params IvyRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Sbt(params SbtRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Helm(params HelmRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Cocoapods(params CocoapodsRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Opkg(params OpkgRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Rpm(params RpmRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Nuget(params NugetRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Cran(params CranRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Gems(params GemsRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Npm(params NpmRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Bower(params BowerRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Debian(params DebianRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Pypi(params PypiRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Docker(params DockerRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Yum(params YumRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Vcs(params VcsRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Composer(params ComposerRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Go(params GoRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) P2(params P2RemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Chef(params ChefRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Puppet(params PuppetRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Conda(params CondaRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Conan(params ConanRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Gitlfs(params GitlfsRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

func (rrs *RemoteRepositoryService) Generic(params GenericRemoteRepositoryParams) error {
	return rrs.performRequest(params, params.Key)
}

type ContentSynchronisation struct {
	Enabled    bool `json:"enables,omitempty"`
	Statistics struct {
		Enabled bool `json:"enables,omitempty"`
	} `json:"statistics,omitempty"`
	Properties struct {
		Enabled bool `json:"enables,omitempty"`
	} `json:"properties,omitempty"`
	Source struct {
		OriginAbsenceDetection bool `json:"originAbsenceDetection,omitempty"`
	} `json:"source,omitempty"`
}

type RemoteRepositoryBaseParams struct {
	Key                               string                  `json:"key,omitempty"`
	Rclass                            string                  `json:"rclass"`
	PackageType                       string                  `json:"packageType,omitempty"`
	Url                               string                  `json:"url"`
	Username                          string                  `json:"username,omitempty"`
	Password                          string                  `json:"password,omitempty"`
	Proxy                             string                  `json:"proxy,omitempty"`
	Description                       string                  `json:"description,omitempty"`
	Notes                             string                  `json:"notes,omitempty"`
	IncludesPattern                   string                  `json:"includesPattern,omitempty"`
	ExcludesPattern                   string                  `json:"excludesPattern,omitempty"`
	RepoLayoutRef                     string                  `json:"repoLayoutRef,omitempty"`
	HardFail                          *bool                   `json:"hardFail,omitempty"`
	Offline                           *bool                   `json:"offline,omitempty"`
	BlackedOut                        *bool                   `json:"blackedOut,omitempty"`
	StoreArtifactsLocally             *bool                   `json:"storeArtifactsLocally,omitempty"`
	SocketTimeoutMillis               int                     `json:"socketTimeoutMillis,omitempty"`
	LocalAddress                      string                  `json:"localAddress,omitempty"`
	RetrievalCachePeriodSecs          int                     `json:"retrievalCachePeriodSecs,omitempty"`
	FailedRetrievalCachePeriodSecs    int                     `json:"failedRetrievalCachePeriodSecs,omitempty"`
	MissedRetrievalCachePeriodSecs    int                     `json:"missedRetrievalCachePeriodSecs,omitempty"`
	UnusedArtifactsCleanupEnabled     *bool                   `json:"unusedArtifactsCleanupEnabled,omitempty"`
	UnusedArtifactsCleanupPeriodHours int                     `json:"unusedArtifactsCleanupPeriodHours,omitempty"`
	AssumedOfflinePeriodSecs          int                     `json:"assumedOfflinePeriodSecs,omitempty"`
	ShareConfiguration                *bool                   `json:"shareConfiguration,omitempty"`
	SynchronizeProperties             *bool                   `json:"synchronizeProperties,omitempty"`
	BlockMismatchingMimeTypes         *bool                   `json:"blockMismatchingMimeTypes,omitempty"`
	PropertySets                      []string                `json:"propertySets,omitempty"`
	AllowAnyHostAuth                  *bool                   `json:"allowAnyHostAuth,omitempty"`
	EnableCookieManagement            *bool                   `json:"enableCookieManagement,omitempty"`
	BypassHeadRequests                *bool                   `json:"bypassHeadRequests,omitempty"`
	ClientTlsCertificate              string                  `json:"clientTlsCertificate,omitempty"`
	BlockPushingSchema1               *bool                   `json:"blockPushingSchema1,omitempty"`
	ContentSynchronisation            *ContentSynchronisation `json:"contentSynchronisation,omitempty"`
}

type CommonMavenGradleRemoteRepositoryParams struct {
	FetchJarsEagerly             *bool  `json:"fetchJarsEagerly,omitempty"`
	FetchSourcesEagerly          *bool  `json:"fetchSourcesEagerly,omitempty"`
	RemoteRepoChecksumPolicyType string `json:"remoteRepoChecksumPolicyType,omitempty"`
	ListRemoteFolderItems        *bool  `json:"listRemoteFolderItems,omitempty"`
	HandleReleases               *bool  `json:"handleReleases,omitempty"`
	HandleSnapshots              *bool  `json:"handleSnapshots,omitempty"`
	SuppressPomConsistencyChecks *bool  `json:"suppressPomConsistencyChecks,omitempty"`
	RejectInvalidJars            *bool  `json:"rejectInvalidJars,omitempty"`
}

type MavenRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
	CommonMavenGradleRemoteRepositoryParams
}

func NewMavenRemoteRepositoryParams() MavenRemoteRepositoryParams {
	return MavenRemoteRepositoryParams{RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "maven"}}
}

type GradleRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
	CommonMavenGradleRemoteRepositoryParams
}

func NewGradleRemoteRepositoryParams() GradleRemoteRepositoryParams {
	return GradleRemoteRepositoryParams{RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "gradle"}}
}

type CocoapodsRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
	PodsSpecsRepoUrl string `json:"podsSpecsRepoUrl,omitempty"`
}

func NewCocoapodsRemoteRepositoryParams() CocoapodsRemoteRepositoryParams {
	return CocoapodsRemoteRepositoryParams{RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "cocoapods"}}
}

type OpkgRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
	ListRemoteFolderItems *bool `json:"listRemoteFolderItems,omitempty"`
}

func NewOpkgRemoteRepositoryParams() OpkgRemoteRepositoryParams {
	return OpkgRemoteRepositoryParams{RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "opkg"}}
}

type RpmRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
	ListRemoteFolderItems *bool `json:"listRemoteFolderItems,omitempty"`
}

func NewRpmRemoteRepositoryParams() RpmRemoteRepositoryParams {
	return RpmRemoteRepositoryParams{RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "rpm"}}
}

type NugetRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
	FeedContextPath          string `json:"feedContextPath,omitempty"`
	DownloadContextPath      string `json:"downloadContextPath,omitempty"`
	V3FeedUrl                string `json:"v3FeedUrl,omitempty"`
	ForceNugetAuthentication *bool  `json:"forceNugetAuthentication,omitempty"`
}

func NewNugetRemoteRepositoryParams() NugetRemoteRepositoryParams {
	return NugetRemoteRepositoryParams{RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "nuget"}}
}

type GemsRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
	ListRemoteFolderItems *bool `json:"listRemoteFolderItems,omitempty"`
}

func NewGemsRemoteRepositoryParams() GemsRemoteRepositoryParams {
	return GemsRemoteRepositoryParams{RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "gems"}}
}

type NpmRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
	ListRemoteFolderItems *bool `json:"listRemoteFolderItems,omitempty"`
}

func NewNpmRemoteRepositoryParams() NpmRemoteRepositoryParams {
	return NpmRemoteRepositoryParams{RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "npm"}}
}

type BowerRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
	BowerRegistryUrl string `json:"bowerRegistryUrl,omitempty"`
}

func NewBowerRemoteRepositoryParams() BowerRemoteRepositoryParams {
	return BowerRemoteRepositoryParams{RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "bower"}}
}

type DebianRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
	ListRemoteFolderItems *bool `json:"listRemoteFolderItems,omitempty"`
}

func NewDebianRemoteRepositoryParams() DebianRemoteRepositoryParams {
	return DebianRemoteRepositoryParams{RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "debian"}}
}

type ComposerRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
	ComposerRegistryUrl string `json:"composerRegistryUrl,omitempty"`
}

func NewComposerRemoteRepositoryParams() ComposerRemoteRepositoryParams {
	return ComposerRemoteRepositoryParams{RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "composer"}}
}

type PypiRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
	ListRemoteFolderItems *bool  `json:"listRemoteFolderItems,omitempty"`
	PypiRegistryUrl       string `json:"pypiRegistryUrl,omitempty"`
}

func NewPypiRemoteRepositoryParams() PypiRemoteRepositoryParams {
	return PypiRemoteRepositoryParams{RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "pypi"}}
}

type DockerRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
	ExternalDependenciesEnabled  *bool    `json:"externalDependenciesEnabled,omitempty"`
	ExternalDependenciesPatterns []string `json:"externalDependenciesPatterns,omitempty"`
	EnableTokenAuthentication    *bool    `json:"enableTokenAuthentication,omitempty"`
}

func NewDockerRemoteRepositoryParams() DockerRemoteRepositoryParams {
	return DockerRemoteRepositoryParams{RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "docker"}}
}

type GitlfsRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
	ListRemoteFolderItems *bool `json:"listRemoteFolderItems,omitempty"`
}

func NewGitlfsRemoteRepositoryParams() GitlfsRemoteRepositoryParams {
	return GitlfsRemoteRepositoryParams{RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "gitlfs"}}
}

type VcsRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
	VcsGitProvider        string `json:"vcsGitProvider,omitempty"`
	VcsType               string `json:"vcsType,omitempty"`
	MaxUniqueSnapshots    int    `json:"maxUniqueSnapshots,omitempty"`
	VcsGitDownloadUrl     string `json:"vcsGitDownloadUrl,omitempty"`
	ListRemoteFolderItems *bool  `json:"listRemoteFolderItems,omitempty"`
}

func NewVcsRemoteRepositoryParams() VcsRemoteRepositoryParams {
	return VcsRemoteRepositoryParams{RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "vcs"}}
}

type GenericRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
	ListRemoteFolderItems *bool `json:"listRemoteFolderItems,omitempty"`
}

func NewGenericRemoteRepositoryParams() GenericRemoteRepositoryParams {
	return GenericRemoteRepositoryParams{RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "generic"}}
}

type IvyRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
}

func NewIvyRemoteRepositoryParams() IvyRemoteRepositoryParams {
	return IvyRemoteRepositoryParams{RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "ivy"}}
}

type SbtRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
}

func NewSbtRemoteRepositoryParams() SbtRemoteRepositoryParams {
	return SbtRemoteRepositoryParams{RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "sbt"}}
}

type HelmRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
}

func NewHelmRemoteRepositoryParams() HelmRemoteRepositoryParams {
	return HelmRemoteRepositoryParams{RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "helm"}}
}

type CranRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
}

func NewCranRemoteRepositoryParams() CranRemoteRepositoryParams {
	return CranRemoteRepositoryParams{RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "cran"}}
}

type GoRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
}

func NewGoRemoteRepositoryParams() GoRemoteRepositoryParams {
	return GoRemoteRepositoryParams{RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "go"}}
}

type YumRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
}

func NewYumRemoteRepositoryParams() YumRemoteRepositoryParams {
	return YumRemoteRepositoryParams{RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "yum"}}
}

type P2RemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
}

func NewP2RemoteRepositoryParams() P2RemoteRepositoryParams {
	return P2RemoteRepositoryParams{RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "p2"}}
}

type ChefRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
}

func NewChefRemoteRepositoryParams() ChefRemoteRepositoryParams {
	return ChefRemoteRepositoryParams{RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "chef"}}
}

type PuppetRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
}

func NewPuppetRemoteRepositoryParams() PuppetRemoteRepositoryParams {
	return PuppetRemoteRepositoryParams{RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "puppet"}}
}

type CondaRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
}

func NewCondaRemoteRepositoryParams() CondaRemoteRepositoryParams {
	return CondaRemoteRepositoryParams{RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "conda"}}
}

type ConanRemoteRepositoryParams struct {
	RemoteRepositoryBaseParams
}

func NewConanRemoteRepositoryParams() ConanRemoteRepositoryParams {
	return ConanRemoteRepositoryParams{RemoteRepositoryBaseParams{Rclass: "remote", PackageType: "conan"}}
}
