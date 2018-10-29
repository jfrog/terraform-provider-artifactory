package artifactory

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type RepositoriesService service

type RepositoryDetails struct {
	Key         *string `json:"key,omitempty"`
	Type        *string `json:"type,omitempty"`
	Description *string `json:"description,omitempty"`
	URL         *string `json:"url,omitempty"`
}

func (r RepositoryDetails) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

type RepositoryListOptions struct {
	// Type of repositories to list.
	// Can be one of local|remote|virtual|distribution. Default: all
	Type string `url:"type,omitempty"`
}

// Returns a list of minimal repository details for all repositories of the specified type.
// Since: 2.2.0
// Security: Requires a privileged user (can be anonymous)
func (s *RepositoriesService) ListRepositories(ctx context.Context, opt *RepositoryListOptions) (*[]RepositoryDetails, *http.Response, error) {
	path := "/api/repositories/"
	req, err := s.client.NewJSONEncodedRequest("GET", path, opt)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", mediaTypeRepositoryDetails)

	repos := new([]RepositoryDetails)
	resp, err := s.client.Do(ctx, req, &repos)
	return repos, resp, err
}

// application/vnd.org.jfrog.artifactory.repositories.LocalRepositoryConfiguration+json
type LocalRepository struct {
	Key                          *string   `json:"key,omitempty"`
	RClass                       *string   `json:"rclass,omitempty"` // Mandatory element in create/replace queries (optional in "update" queries)
	PackageType                  *string   `json:"packageType,omitempty"`
	Description                  *string   `json:"description,omitempty"`
	Notes                        *string   `json:"notes,omitempty"`
	IncludesPattern              *string   `json:"includesPattern,omitempty"`
	ExcludesPattern              *string   `json:"excludesPattern,omitempty"`
	ArchiveBrowsingEnabled       *bool     `json:"archiveBrowsingEnabled,omitempty"`
	BlackedOut                   *bool     `json:"blackedOut,omitempty"`
	BlockXrayUnscannedArtifacts  *bool     `json:"blockXrayUnscannedArtifacts,omitempty"`
	CalculateYumMetadata         *bool     `json:"calculateYumMetadata,omitempty"`
	ChecksumPolicyType           *string   `json:"checksumPolicyType,omitempty"`
	DebianTrivialLayout          *bool     `json:"debianTrivialLayout,omitempty"`
	DockerApiVersion             *string   `json:"dockerApiVersion,omitempty"`
	EnableBowerSupport           *bool     `json:"enableBowerSupport,omitempty"`
	EnableCocoaPodsSupport       *bool     `json:"enableCocoaPodsSupport,omitempty"`
	EnableComposerSupport        *bool     `json:"enableComposerSupport,omitempty"`
	EnableConanSupport           *bool     `json:"enableConanSupport,omitempty"`
	EnableDebianSupport          *bool     `json:"enableDebianSupport,omitempty"`
	EnableDistRepoSupport        *bool     `json:"enableDistRepoSupport,omitempty"`
	EnableDockerSupport          *bool     `json:"enableDockerSupport,omitempty"`
	EnableFileListsIndexing      *bool     `json:"enableFileListsIndexing,omitempty"`
	EnableGemsSupport            *bool     `json:"enableGemsSupport,omitempty"`
	EnableGitLfsSupport          *bool     `json:"enableGitLfsSupport,omitempty"`
	EnableNpmSupport             *bool     `json:"enableNpmSupport,omitempty"`
	EnableNuGetSupport           *bool     `json:"enableNuGetSupport,omitempty"`
	EnablePuppetSupport          *bool     `json:"enablePuppetSupport,omitempty"`
	EnablePypiSupport            *bool     `json:"enablePypiSupport,omitempty"`
	EnableVagrantSupport         *bool     `json:"enableVagrantSupport,omitempty"`
	EnabledChefSupport           *bool     `json:"enabledChefSupport,omitempty"`
	ForceNugetAuthentication     *bool     `json:"forceNugetAuthentication,omitempty"`
	HandleReleases               *bool     `json:"handleReleases,omitempty"`
	HandleSnapshots              *bool     `json:"handleSnapshots,omitempty"`
	MaxUniqueSnapshots           *int      `json:"maxUniqueSnapshots,omitempty"`
	MaxUniqueTags                *int      `json:"maxUniqueTags,omitempty"`
	PropertySets                 *[]string `json:"propertySets,omitempty"`
	RepoLayoutRef                *string   `json:"repoLayoutRef,omitempty"`
	SnapshotVersionBehavior      *string   `json:"snapshotVersionBehavior,omitempty"`
	SuppressPomConsistencyChecks *bool     `json:"suppressPomConsistencyChecks,omitempty"`
	XrayIndex                    *bool     `json:"xrayIndex,omitempty"`
	XrayMinimumBlockedSeverity   *string   `json:"xrayMinimumBlockedSeverity,omitempty"`
	YumRootDepth                 *int      `json:"yumRootDepth,omitempty"`
}

func (r LocalRepository) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

// Creates a new repository in Artifactory with the provided configuration.
// Since: 2.3.0
// Notes: Requires Artifactory Pro
// An existing repository with the same key are removed from the configuration and its content is removed!
// Missing values are set to the default values as defined by the consumed type spec.
// Security: Requires an admin user
func (s *RepositoriesService) CreateLocal(ctx context.Context, repo *LocalRepository) (*http.Response, error) {
	return s.create(ctx, *repo.Key, repo)
}

// Retrieves the current configuration of a repository.
// Since: 2.3.0
// Notes: Requires Artifactory Pro
// Security: Requires an admin user for complete repository configuration. Non-admin users will receive only partial configuration data.
func (s *RepositoriesService) GetLocal(ctx context.Context, repo string) (*LocalRepository, *http.Response, error) {
	repository, resp, err := s.get(ctx, repo, new(LocalRepository))
	if err != nil {
		return nil, resp, err
	}
	return repository.(*LocalRepository), resp, nil
}

// Updates an exiting repository configuration in Artifactory with the provided configuration elements.
// Since: 2.3.0
// Notes: Requires Artifactory Pro
// The class of a repository (the rclass attribute cannot be updated.
// Security: Requires an admin user
func (s *RepositoriesService) UpdateLocal(ctx context.Context, repo string, repository *LocalRepository) (*http.Response, error) {
	return s.update(ctx, repo, repository)
}

// Removes a repository configuration together with the whole repository content.
// Since: 2.3.0
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *RepositoriesService) DeleteLocal(ctx context.Context, repo string) (*http.Response, error) {
	return s.delete(ctx, repo)
}

type Statistics struct {
	Enabled *bool `json:"enabled,omitempty"`
}

type Properties struct {
	Enabled *bool `json:"enabled,omitempty"`
}

type Source struct {
	OriginAbsenceDetection *bool `json:"originAbsenceDetection,omitempty"`
}

type ContentSynchronisation struct {
	Enabled    *bool       `json:"enabled,omitempty"`
	Statistics *Statistics `json:"statistics,omitempty"`
	Properties *Properties `json:"properties,omitempty"`
	Source     *Source     `json:"source,omitempty"`
}

type RemoteRepository struct {
	Key                               *string                 `json:"key,omitempty"`
	RClass                            *string                 `json:"rclass,omitempty"` // Mandatory element in create/replace queries (optional in "update" queries)
	PackageType                       *string                 `json:"packageType,omitempty"`
	Description                       *string                 `json:"description,omitempty"`
	Notes                             *string                 `json:"notes,omitempty"`
	IncludesPattern                   *string                 `json:"includesPattern,omitempty"`
	ExcludesPattern                   *string                 `json:"excludesPattern,omitempty"`
	AllowAnyHostAuth                  *bool                   `json:"allowAnyHostAuth,omitempty"`
	ArchiveBrowsingEnabled            *bool                   `json:"archiveBrowsingEnabled,omitempty"`
	AssumedOfflinePeriodSecs          *int                    `json:"assumedOfflinePeriodSecs,omitempty"`
	BlackedOut                        *bool                   `json:"blackedOut,omitempty"`
	BlockMismatchingMimeTypes         *bool                   `json:"blockMismatchingMimeTypes,omitempty"`
	BlockXrayUnscannedArtifacts       *bool                   `json:"blockXrayUnscannedArtifacts,omitempty"`
	BypassHeadRequests                *bool                   `json:"bypassHeadRequests,omitempty"`
	ContentSynchronisation            *ContentSynchronisation `json:"contentSynchronisation,omitempty"`
	DebianTrivialLayout               *bool                   `json:"debianTrivialLayout,omitempty"`
	DockerApiVersion                  *string                 `json:"dockerApiVersion,omitempty"`
	EnableBowerSupport                *bool                   `json:"enableBowerSupport,omitempty"`
	EnableCocoaPodsSupport            *bool                   `json:"enableCocoaPodsSupport,omitempty"`
	EnableConanSupport                *bool                   `json:"enableConanSupport,omitempty"`
	EnableCookieManagement            *bool                   `json:"enableCookieManagement,omitempty"`
	EnabledChefSupport                *bool                   `json:"enabledChefSupport,omitempty"`
	EnableComposerSupport             *bool                   `json:"enableComposerSupport,omitempty"`
	EnableDebianSupport               *bool                   `json:"enableDebianSupport,omitempty"`
	EnableDistRepoSupport             *bool                   `json:"enableDistRepoSupport,omitempty"`
	EnableDockerSupport               *bool                   `json:"enableDockerSupport,omitempty"`
	EnableGemsSupport                 *bool                   `json:"enableGemsSupport,omitempty"`
	EnableGitLfsSupport               *bool                   `json:"enableGitLfsSupport,omitempty"`
	EnableNpmSupport                  *bool                   `json:"enableNpmSupport,omitempty"`
	EnableNuGetSupport                *bool                   `json:"enableNuGetSupport,omitempty"`
	EnablePuppetSupport               *bool                   `json:"enablePuppetSupport,omitempty"`
	EnablePypiSupport                 *bool                   `json:"enablePypiSupport,omitempty"`
	EnableTokenAuthentication         *bool                   `json:"enableTokenAuthentication,omitempty"`
	EnableVagrantSupport              *bool                   `json:"enableVagrantSupport,omitempty"`
	FailedRetrievalCachePeriodSecs    *int                    `json:"failedRetrievalCachePeriodSecs,omitempty"`
	FetchJarsEagerly                  *bool                   `json:"fetchJarsEagerly,omitempty"`
	FetchSourcesEagerly               *bool                   `json:"fetchSourcesEagerly,omitempty"`
	ForceNugetAuthentication          *bool                   `json:"forceNugetAuthentication,omitempty"`
	HandleReleases                    *bool                   `json:"handleReleases,omitempty"`
	HandleSnapshots                   *bool                   `json:"handleSnapshots,omitempty"`
	HardFail                          *bool                   `json:"hardFail,omitempty"`
	ListRemoteFolderItems             *bool                   `json:"listRemoteFolderItems,omitempty"`
	LocalAddress                      *string                 `json:"localAddress,omitempty"`
	MaxUniqueSnapshots                *int                    `json:"maxUniqueSnapshots,omitempty"`
	MaxUniqueTags                     *int                    `json:"maxUniqueTags,omitempty"`
	MismatchingMimeTypesOverrideList  *string                 `json:"mismatchingMimeTypesOverrideList,omitempty"`
	MissedRetrievalCachePeriodSecs    *int                    `json:"missedRetrievalCachePeriodSecs,omitempty"`
	Offline                           *bool                   `json:"offline,omitempty"`
	Password                          *string                 `json:"password,omitempty"`
	PropagateQueryParams              *bool                   `json:"propagateQueryParams,omitempty"`
	PropertySets                      *[]string               `json:"propertySets,omitempty"`
	Proxy                             *string                 `json:"proxy,omitempty"`
	RejectInvalidJars                 *bool                   `json:"rejectInvalidJars,omitempty"`
	RemoteRepoChecksumPolicyType      *string                 `json:"remoteRepoChecksumPolicyType,omitempty"`
	RepoLayoutRef                     *string                 `json:"repoLayoutRef,omitempty"`
	RetrievalCachePeriodSecs          *int                    `json:"retrievalCachePeriodSecs,omitempty"`
	ShareConfiguration                *bool                   `json:"shareConfiguration,omitempty"`
	SocketTimeoutMillis               *int                    `json:"socketTimeoutMillis,omitempty"`
	StoreArtifactsLocally             *bool                   `json:"storeArtifactsLocally,omitempty"`
	SuppressPomConsistencyChecks      *bool                   `json:"suppressPomConsistencyChecks,omitempty"`
	SynchronizeProperties             *bool                   `json:"synchronizeProperties,omitempty"`
	UnusedArtifactsCleanupEnabled     *bool                   `json:"unusedArtifactsCleanupEnabled,omitempty"`
	UnusedArtifactsCleanupPeriodHours *int                    `json:"unusedArtifactsCleanupPeriodHours,omitempty"`
	Url                               *string                 `json:"url,omitempty"`
	Username                          *string                 `json:"username,omitempty"` // Mandatory element in create/replace queries (optional in "update" queries)
	XrayIndex                         *bool                   `json:"xrayIndex,omitempty"`
	XrayMinimumBlockedSeverity        *string                 `json:"xrayMinimumBlockedSeverity,omitempty"`
	BowerRegistryURL                  *string                 `json:"bowerRegistryUrl,omitempty"`
	VcsType                           *string                 `json:"vcsType,omitempty"`
	VcsGitProvider                    *string                 `json:"vcsGitProvider,omitempty"`
	VcsGitDownloadUrl                 *string                 `json:"vcsGitDownloadUrl,omitempty"`
	ClientTLSCertificate              *string                 `json:"clientTlsCertificate,omitempty"`
	PyPiRegistryUrl                   *string                 `json:"pyPiRegistryUrl,omitempty"`
}

func (r RemoteRepository) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

// Creates a new repository in Artifactory with the provided configuration.
// Since: 2.3.0
// Notes: Requires Artifactory Pro
// An existing repository with the same key are removed from the configuration and its content is removed!
// Missing values are set to the default values as defined by the consumed type spec.
// Security: Requires an admin user
func (s *RepositoriesService) CreateRemote(ctx context.Context, repo *RemoteRepository) (*http.Response, error) {
	return s.create(ctx, *repo.Key, repo)
}

// Retrieves the current configuration of a repository.
// Since: 2.3.0
// Notes: Requires Artifactory Pro
// Security: Requires an admin user for complete repository configuration. Non-admin users will receive only partial configuration data.
func (s *RepositoriesService) GetRemote(ctx context.Context, repo string) (*RemoteRepository, *http.Response, error) {
	repository, resp, err := s.get(ctx, repo, new(RemoteRepository))
	if err != nil {
		return nil, resp, err
	}
	return repository.(*RemoteRepository), resp, nil
}

// Updates an exiting repository configuration in Artifactory with the provided configuration elements.
// Since: 2.3.0
// Notes: Requires Artifactory Pro
// The class of a repository (the rclass attribute cannot be updated.
// Security: Requires an admin user
func (s *RepositoriesService) UpdateRemote(ctx context.Context, repo string, repository *RemoteRepository) (*http.Response, error) {
	return s.update(ctx, repo, repository)
}

// Removes a repository configuration together with the whole repository content.
// Since: 2.3.0
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *RepositoriesService) DeleteRemote(ctx context.Context, repo string) (*http.Response, error) {
	return s.delete(ctx, repo)
}

type VirtualRepository struct {
	Key                                           *string   `json:"key,omitempty"`
	RClass                                        *string   `json:"rclass,omitempty"` // Mandatory element in create/replace queries (optional in "update" queries)
	PackageType                                   *string   `json:"packageType,omitempty"`
	Description                                   *string   `json:"description,omitempty"`
	Notes                                         *string   `json:"notes,omitempty"`
	IncludesPattern                               *string   `json:"includesPattern,omitempty"`
	ExcludesPattern                               *string   `json:"excludesPattern,omitempty"`
	ArtifactoryRequestsCanRetrieveRemoteArtifacts *bool     `json:"artifactoryRequestsCanRetrieveRemoteArtifacts,omitempty"`
	DebianTrivialLayout                           *bool     `json:"debianTrivialLayout,omitempty"`
	DefaultDeploymentRepo                         *string   `json:"defaultDeploymentRepo,omitempty"`
	DockerApiVersion                              *string   `json:"dockerApiVersion,omitempty"`
	EnableBowerSupport                            *bool     `json:"enableBowerSupport,omitempty"`
	EnableCocoaPodsSupport                        *bool     `json:"enableCocoaPodsSupport,omitempty"`
	EnableConanSupport                            *bool     `json:"enableConanSupport,omitempty"`
	EnableComposerSupport                         *bool     `json:"enableComposerSupport,omitempty"`
	EnabledChefSupport                            *bool     `json:"enabledChefSupport,omitempty"`
	EnableDebianSupport                           *bool     `json:"enableDebianSupport,omitempty"`
	EnableDistRepoSupport                         *bool     `json:"enableDistRepoSupport,omitempty"`
	EnableDockerSupport                           *bool     `json:"enableDockerSupport,omitempty"`
	EnableGemsSupport                             *bool     `json:"enableGemsSupport,omitempty"`
	EnableGitLfsSupport                           *bool     `json:"enableGitLfsSupport,omitempty"`
	EnableNpmSupport                              *bool     `json:"enableNpmSupport,omitempty"`
	EnableNuGetSupport                            *bool     `json:"enableNuGetSupport,omitempty"`
	EnablePuppetSupport                           *bool     `json:"enablePuppetSupport,omitempty"`
	EnablePypiSupport                             *bool     `json:"enablePypiSupport,omitempty"`
	EnableVagrantSupport                          *bool     `json:"enableVagrantSupport,omitempty"`
	ExternalDependenciesEnabled                   *bool     `json:"externalDependenciesEnabled,omitempty"`
	ForceNugetAuthentication                      *bool     `json:"forceNugetAuthentication,omitempty"`
	KeyPair                                       *string   `json:"keyPair,omitempty"`
	PomRepositoryReferencesCleanupPolicy          *string   `json:"pomRepositoryReferencesCleanupPolicy,omitempty"`
	Repositories                                  *[]string `json:"repositories,omitempty"`
	VirtualRetrievalCachePeriodSecs               *int      `json:"virtualRetrievalCachePeriodSecs,omitempty"`
}

func (r VirtualRepository) String() string {
	res, _ := json.MarshalIndent(r, "", "    ")
	return string(res)
}

// Creates a new repository in Artifactory with the provided configuration.
// Since: 2.3.0
// Notes: Requires Artifactory Pro
// An existing repository with the same key are removed from the configuration and its content is removed!
// Missing values are set to the default values as defined by the consumed type spec.
// Security: Requires an admin user
func (s *RepositoriesService) CreateVirtual(ctx context.Context, repo *VirtualRepository) (*http.Response, error) {
	return s.create(ctx, *repo.Key, repo)
}

// Retrieves the current configuration of a repository.
// Since: 2.3.0
// Notes: Requires Artifactory Pro
// Security: Requires an admin user for complete repository configuration. Non-admin users will receive only partial configuration data.
func (s *RepositoriesService) GetVirtual(ctx context.Context, repo string) (*VirtualRepository, *http.Response, error) {
	repository, resp, err := s.get(ctx, repo, new(VirtualRepository))
	if err != nil {
		return nil, resp, err
	}
	return repository.(*VirtualRepository), resp, nil
}

// Updates an exiting repository configuration in Artifactory with the provided configuration elements.
// Since: 2.3.0
// Notes: Requires Artifactory Pro
// The class of a repository (the rclass attribute cannot be updated.
// Security: Requires an admin user
func (s *RepositoriesService) UpdateVirtual(ctx context.Context, repo string, repository *VirtualRepository) (*http.Response, error) {
	return s.update(ctx, repo, repository)
}

// Removes a repository configuration together with the whole repository content.
// Since: 2.3.0
// Notes: Requires Artifactory Pro
// Security: Requires an admin user
func (s *RepositoriesService) DeleteVirtual(ctx context.Context, repo string) (*http.Response, error) {
	return s.delete(ctx, repo)
}

// Generic repo CRUD operations
func (s *RepositoriesService) create(ctx context.Context, repo string, v interface{}) (*http.Response, error) {
	path := fmt.Sprintf("/api/repositories/%s", repo)
	req, err := s.client.NewJSONEncodedRequest("PUT", path, v)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

func (s *RepositoriesService) get(ctx context.Context, repo string, v interface{}) (interface{}, *http.Response, error) {
	path := fmt.Sprintf("/api/repositories/%v", repo)
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	acceptHeaders := []string{mediaTypeLocalRepository, mediaTypeVirtualRepository, mediaTypeRemoteRepository}
	req.Header.Set("Accept", strings.Join(acceptHeaders, ", "))

	resp, err := s.client.Do(ctx, req, v)
	return v, resp, err
}

func (s *RepositoriesService) update(ctx context.Context, repo string, v interface{}) (*http.Response, error) {
	path := fmt.Sprintf("/api/repositories/%v", repo)
	req, err := s.client.NewJSONEncodedRequest("POST", path, v)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

func (s *RepositoriesService) delete(ctx context.Context, repo string) (*http.Response, error) {
	path := fmt.Sprintf("/api/repositories/%v", repo)
	req, err := s.client.NewJSONEncodedRequest("DELETE", path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}
