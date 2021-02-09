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

type LocalRepositoryService struct {
	isUpdate   bool
	client     *rthttpclient.ArtifactoryHttpClient
	ArtDetails auth.ServiceDetails
}

func NewLocalRepositoryService(client *rthttpclient.ArtifactoryHttpClient, isUpdate bool) *LocalRepositoryService {
	return &LocalRepositoryService{client: client, isUpdate: isUpdate}
}

func (lrs *LocalRepositoryService) GetJfrogHttpClient() *rthttpclient.ArtifactoryHttpClient {
	return lrs.client
}

func (lrs *LocalRepositoryService) performRequest(params interface{}, repoKey string) error {
	content, err := json.Marshal(params)
	if errorutils.CheckError(err) != nil {
		return err
	}
	httpClientsDetails := lrs.ArtDetails.CreateHttpClientDetails()
	utils.SetContentType("application/vnd.org.jfrog.artifactory.repositories.LocalRepositoryConfiguration+json", &httpClientsDetails.Headers)
	var url = lrs.ArtDetails.GetUrl() + "api/repositories/" + repoKey
	var operationString string
	var resp *http.Response
	var body []byte
	if lrs.isUpdate {
		log.Info("Updating local repository...")
		operationString = "updating"
		resp, body, err = lrs.client.SendPost(url, content, &httpClientsDetails)
	} else {
		log.Info("Creating local repository...")
		operationString = "creating"
		resp, body, err = lrs.client.SendPut(url, content, &httpClientsDetails)
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

func (lrs *LocalRepositoryService) Maven(params MavenLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Gradle(params GradleLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Rpm(params RpmLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Nuget(params NugetLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Debian(params DebianLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Docker(params DockerLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Ivy(params IvyLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Sbt(params SbtLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Helm(params HelmLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Cocoapods(params CocoapodsLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Opkg(params OpkgLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Cran(params CranLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Gems(params GemsLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Npm(params NpmLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Bower(params BowerLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Composer(params ComposerLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Pypi(params PypiLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Vagrant(params VagrantLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Gitlfs(params GitlfsLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Go(params GoLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Yum(params YumLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Conan(params ConanLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Chef(params ChefLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Puppet(params PuppetLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

func (lrs *LocalRepositoryService) Generic(params GenericLocalRepositoryParams) error {
	return lrs.performRequest(params, params.Key)
}

type LocalRepositoryBaseParams struct {
	Key                             string   `json:"key,omitempty"`
	Rclass                          string   `json:"rclass"`
	PackageType                     string   `json:"packageType,omitempty"`
	Description                     string   `json:"description,omitempty"`
	Notes                           string   `json:"notes,omitempty"`
	IncludesPattern                 string   `json:"includesPattern,omitempty"`
	ExcludesPattern                 string   `json:"excludesPattern,omitempty"`
	RepoLayoutRef                   string   `json:"repoLayoutRef,omitempty"`
	BlackedOut                      *bool    `json:"blackedOut,omitempty"`
	XrayIndex                       *bool    `json:"xrayIndex,omitempty"`
	PropertySets                    []string `json:"propertySets,omitempty"`
	ArchiveBrowsingEnabled          *bool    `json:"archiveBrowsingEnabled,omitempty"`
	OptionalIndexCompressionFormats []string `json:"optionalIndexCompressionFormats,omitempty"`
	DownloadRedirect                *bool    `json:"downloadRedirect,omitempty"`
	BlockPushingSchema1             *bool    `json:"blockPushingSchema1,omitempty"`
}

type CommonMavenGradleLocalRepositoryParams struct {
	MaxUniqueSnapshots           int    `json:"maxUniqueSnapshots,omitempty"`
	HandleReleases               *bool  `json:"handleReleases,omitempty"`
	HandleSnapshots              *bool  `json:"handleSnapshots,omitempty"`
	SuppressPomConsistencyChecks *bool  `json:"suppressPomConsistencyChecks,omitempty"`
	SnapshotVersionBehavior      string `json:"snapshotVersionBehavior,omitempty"`
	ChecksumPolicyType           string `json:"checksumPolicyType,omitempty"`
}

type MavenLocalRepositoryParams struct {
	LocalRepositoryBaseParams
	CommonMavenGradleLocalRepositoryParams
}

func NewMavenLocalRepositoryParams() MavenLocalRepositoryParams {
	return MavenLocalRepositoryParams{LocalRepositoryBaseParams: LocalRepositoryBaseParams{Rclass: "local", PackageType: "maven"}}
}

type GradleLocalRepositoryParams struct {
	LocalRepositoryBaseParams
	CommonMavenGradleLocalRepositoryParams
}

func NewGradleLocalRepositoryParams() GradleLocalRepositoryParams {
	return GradleLocalRepositoryParams{LocalRepositoryBaseParams: LocalRepositoryBaseParams{Rclass: "local", PackageType: "gradle"}}
}

type RpmLocalRepositoryParams struct {
	LocalRepositoryBaseParams
	YumRootDepth            int   `json:"yumRootDepth,omitempty"`
	CalculateYumMetadata    *bool `json:"calculateYumMetadata,omitempty"`
	EnableFileListsIndexing *bool `json:"enableFileListsIndexing,omitempty"`
}

func NewRpmLocalRepositoryParams() RpmLocalRepositoryParams {
	return RpmLocalRepositoryParams{LocalRepositoryBaseParams: LocalRepositoryBaseParams{Rclass: "local", PackageType: "rpm"}}
}

type NugetLocalRepositoryParams struct {
	LocalRepositoryBaseParams
	MaxUniqueSnapshots       int   `json:"maxUniqueSnapshots,omitempty"`
	ForceNugetAuthentication *bool `json:"forceNugetAuthentication,omitempty"`
}

func NewNugetLocalRepositoryParams() NugetLocalRepositoryParams {
	return NugetLocalRepositoryParams{LocalRepositoryBaseParams: LocalRepositoryBaseParams{Rclass: "local", PackageType: "nuget"}}
}

type DebianLocalRepositoryParams struct {
	LocalRepositoryBaseParams
	DebianTrivialLayout *bool `json:"debianTrivialLayout,omitempty"`
}

func NewDebianLocalRepositoryParams() DebianLocalRepositoryParams {
	return DebianLocalRepositoryParams{LocalRepositoryBaseParams: LocalRepositoryBaseParams{Rclass: "local", PackageType: "debian"}}
}

type DockerLocalRepositoryParams struct {
	LocalRepositoryBaseParams
	MaxUniqueTags    int    `json:"maxUniqueTags,omitempty"`
	DockerApiVersion string `json:"dockerApiVersion,omitempty"`
}

func NewDockerLocalRepositoryParams() DockerLocalRepositoryParams {
	return DockerLocalRepositoryParams{LocalRepositoryBaseParams: LocalRepositoryBaseParams{Rclass: "local", PackageType: "docker"}}
}

type IvyLocalRepositoryParams struct {
	LocalRepositoryBaseParams
}

func NewIvyLocalRepositoryParams() IvyLocalRepositoryParams {
	return IvyLocalRepositoryParams{LocalRepositoryBaseParams{Rclass: "local", PackageType: "ivy"}}
}

type SbtLocalRepositoryParams struct {
	LocalRepositoryBaseParams
}

func NewSbtLocalRepositoryParams() SbtLocalRepositoryParams {
	return SbtLocalRepositoryParams{LocalRepositoryBaseParams{Rclass: "local", PackageType: "sbt"}}
}

type HelmLocalRepositoryParams struct {
	LocalRepositoryBaseParams
}

func NewHelmLocalRepositoryParams() HelmLocalRepositoryParams {
	return HelmLocalRepositoryParams{LocalRepositoryBaseParams{Rclass: "local", PackageType: "helm"}}
}

type CocoapodsLocalRepositoryParams struct {
	LocalRepositoryBaseParams
}

func NewCocoapodsLocalRepositoryParams() CocoapodsLocalRepositoryParams {
	return CocoapodsLocalRepositoryParams{LocalRepositoryBaseParams{Rclass: "local", PackageType: "cocoapods"}}
}

type OpkgLocalRepositoryParams struct {
	LocalRepositoryBaseParams
}

func NewOpkgLocalRepositoryParams() OpkgLocalRepositoryParams {
	return OpkgLocalRepositoryParams{LocalRepositoryBaseParams{Rclass: "local", PackageType: "opkg"}}
}

type CranLocalRepositoryParams struct {
	LocalRepositoryBaseParams
}

func NewCranLocalRepositoryParams() CranLocalRepositoryParams {
	return CranLocalRepositoryParams{LocalRepositoryBaseParams{Rclass: "local", PackageType: "cran"}}
}

type GemsLocalRepositoryParams struct {
	LocalRepositoryBaseParams
}

func NewGemsLocalRepositoryParams() GemsLocalRepositoryParams {
	return GemsLocalRepositoryParams{LocalRepositoryBaseParams{Rclass: "local", PackageType: "gems"}}
}

type NpmLocalRepositoryParams struct {
	LocalRepositoryBaseParams
}

func NewNpmLocalRepositoryParams() NpmLocalRepositoryParams {
	return NpmLocalRepositoryParams{LocalRepositoryBaseParams{Rclass: "local", PackageType: "npm"}}
}

type BowerLocalRepositoryParams struct {
	LocalRepositoryBaseParams
}

func NewBowerLocalRepositoryParams() BowerLocalRepositoryParams {
	return BowerLocalRepositoryParams{LocalRepositoryBaseParams{Rclass: "local", PackageType: "bower"}}
}

type ComposerLocalRepositoryParams struct {
	LocalRepositoryBaseParams
}

func NewComposerLocalRepositoryParams() ComposerLocalRepositoryParams {
	return ComposerLocalRepositoryParams{LocalRepositoryBaseParams{Rclass: "local", PackageType: "composer"}}
}

type PypiLocalRepositoryParams struct {
	LocalRepositoryBaseParams
}

func NewPypiLocalRepositoryParams() PypiLocalRepositoryParams {
	return PypiLocalRepositoryParams{LocalRepositoryBaseParams{Rclass: "local", PackageType: "pypi"}}
}

type VagrantLocalRepositoryParams struct {
	LocalRepositoryBaseParams
}

func NewVagrantLocalRepositoryParams() VagrantLocalRepositoryParams {
	return VagrantLocalRepositoryParams{LocalRepositoryBaseParams{Rclass: "local", PackageType: "vagrant"}}
}

type GitlfsLocalRepositoryParams struct {
	LocalRepositoryBaseParams
}

func NewGitlfsLocalRepositoryParams() GitlfsLocalRepositoryParams {
	return GitlfsLocalRepositoryParams{LocalRepositoryBaseParams{Rclass: "local", PackageType: "gitlfs"}}
}

type GoLocalRepositoryParams struct {
	LocalRepositoryBaseParams
}

func NewGoLocalRepositoryParams() GoLocalRepositoryParams {
	return GoLocalRepositoryParams{LocalRepositoryBaseParams{Rclass: "local", PackageType: "go"}}
}

type YumLocalRepositoryParams struct {
	LocalRepositoryBaseParams
}

func NewYumLocalRepositoryParams() YumLocalRepositoryParams {
	return YumLocalRepositoryParams{LocalRepositoryBaseParams{Rclass: "local", PackageType: "yum"}}
}

type ConanLocalRepositoryParams struct {
	LocalRepositoryBaseParams
}

func NewConanLocalRepositoryParams() ConanLocalRepositoryParams {
	return ConanLocalRepositoryParams{LocalRepositoryBaseParams{Rclass: "local", PackageType: "conan"}}
}

type ChefLocalRepositoryParams struct {
	LocalRepositoryBaseParams
}

func NewChefLocalRepositoryParams() ChefLocalRepositoryParams {
	return ChefLocalRepositoryParams{LocalRepositoryBaseParams{Rclass: "local", PackageType: "chef"}}
}

type PuppetLocalRepositoryParams struct {
	LocalRepositoryBaseParams
}

func NewPuppetLocalRepositoryParams() PuppetLocalRepositoryParams {
	return PuppetLocalRepositoryParams{LocalRepositoryBaseParams{Rclass: "local", PackageType: "puppet"}}
}

type GenericLocalRepositoryParams struct {
	LocalRepositoryBaseParams
}

func NewGenericLocalRepositoryParams() GenericLocalRepositoryParams {
	return GenericLocalRepositoryParams{LocalRepositoryBaseParams{Rclass: "local", PackageType: "generic"}}
}
