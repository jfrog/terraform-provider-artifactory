package local

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	datasource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	"github.com/samber/lo"
)

// LocalDataSourceAttributes defines the attributes for local repository datasources
var LocalDataSourceAttributes = lo.Assign(
	datasource_repository.BaseDataSourceAttributes,
	map[string]schema.Attribute{
		"blacked_out": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "When set, the repository does not participate in artifact resolution and new artifacts cannot be deployed.",
		},
		"xray_index": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Enable Indexing In Xray. Repository will be indexed with the default retention period.",
		},
		"property_sets": schema.SetAttribute{
			ElementType:         types.StringType,
			Computed:            true,
			MarkdownDescription: "List of property set name",
		},
		"archive_browsing_enabled": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "When set, you may view content such as HTML or Javadoc files directly from Artifactory.",
		},
		"download_direct": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "When set, download requests to this repository will redirect the client to download the artifact directly from the cloud storage provider.",
		},
		"priority_resolution": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "Setting repositories with priority will cause metadata to be merged only from repositories set with this field",
		},
		"cdn_redirect": schema.BoolAttribute{
			Computed:            true,
			MarkdownDescription: "When set, download requests to this repository will redirect the client to download the artifact directly from AWS CloudFront.",
		},
	},
)
