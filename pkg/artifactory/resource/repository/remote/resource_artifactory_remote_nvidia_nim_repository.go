package remote

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const NvidiaNimPackageType = "nvidia-nim"

func NewRemoteNvidiaNimRepositoryResource() resource.Resource {
	return &RemoteRepositoryResource[RemoteRepositoryResourceModel]{
		PackageType: NvidiaNimPackageType,
	}
}