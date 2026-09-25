package local

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const NvidiaNimPackageType = "nvidia-nim"

func NewLocalNvidiaNimRepositoryResource() resource.Resource {
	return &LocalRepositoryResource[LocalRepositoryResourceModel]{
		PackageType: NvidiaNimPackageType,
	}
}