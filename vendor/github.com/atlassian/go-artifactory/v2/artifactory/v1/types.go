package v1

import "github.com/atlassian/go-artifactory/v2/artifactory/client"

const (
	mediaTypeLocalRepository   = "application/vnd.org.jfrog.artifactory.repositories.LocalRepositoryConfiguration+json"
	mediaTypeRemoteRepository  = "application/vnd.org.jfrog.artifactory.repositories.RemoteRepositoryConfiguration+json"
	mediaTypeVirtualRepository = "application/vnd.org.jfrog.artifactory.repositories.VirtualRepositoryConfiguration+json"
	mediaTypeRepositoryDetails = "application/vnd.org.jfrog.artifactory.repositories.RepositoryDetailsList+json"
	mediaTypeSystemVersion     = "application/vnd.org.jfrog.artifactory.system.Version+json"
	mediaTypeUsers             = "application/vnd.org.jfrog.artifactory.security.Users+json"
	mediaTypeUser              = "application/vnd.org.jfrog.artifactory.security.User+json"
	mediaTypeGroups            = "application/vnd.org.jfrog.artifactory.security.Groups+json"
	mediaTypeGroup             = "application/vnd.org.jfrog.artifactory.security.Group+json"
	mediaTypePermissionTargets = "application/vnd.org.jfrog.artifactory.security.PermissionTargets+json"
	mediaTypePermissionTarget  = "application/vnd.org.jfrog.artifactory.security.PermissionTarget+json"
	mediaTypeItemPermissions   = "application/vnd.org.jfrog.artifactory.storage.ItemPermissions+json"
	mediaTypeReplicationConfig = "application/vnd.org.jfrog.artifactory.replications.ReplicationConfigRequest+json"
	mediaTypeFileInfo          = "application/vnd.org.jfrog.artifactory.storage.FileInfo+json"
)

type Service struct {
	client *client.Client
}

type V1 struct {
	common Service

	// Services used for talking to different parts of the Artifactory API.
	Repositories *RepositoriesService
	Security     *SecurityService
	System       *SystemService
	Artifacts    *ArtifactService
}
