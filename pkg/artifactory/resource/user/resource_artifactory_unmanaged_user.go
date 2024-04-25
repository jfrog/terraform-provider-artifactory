package user

import "github.com/hashicorp/terraform-plugin-framework/resource"

func NewUnmanagedUserResource() resource.Resource {
	return &ArtifactoryUserResource{
		ArtifactoryBaseUserResource: ArtifactoryBaseUserResource{
			TypeName: "artifactory_unmanaged_user",
		},
	}
}
