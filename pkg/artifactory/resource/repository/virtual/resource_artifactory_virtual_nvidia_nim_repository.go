package virtual

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const NvidiaNimPackageType = "nvidia-nim"

func NewVirtualNvidiaNimRepositoryResource() resource.Resource {
	return &VirtualRepositoryResource[VirtualRepositoryResourceModel]{
		PackageType: NvidiaNimPackageType,
	}
}