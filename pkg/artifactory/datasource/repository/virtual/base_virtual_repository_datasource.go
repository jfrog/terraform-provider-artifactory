package virtual

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	"github.com/samber/lo"
)

// BaseVirtualRepositoryDataSourceModel contains common fields for all virtual repository data sources
type BaseVirtualRepositoryDataSourceModel struct {
	repository.BaseRepositoryDataSourceModel
	Repositories                                  types.Set    `tfsdk:"repositories"`
	ArtifactoryRequestsCanRetrieveRemoteArtifacts types.Bool   `tfsdk:"artifactory_requests_can_retrieve_remote_artifacts"`
	DefaultDeploymentRepo                         types.String `tfsdk:"default_deployment_repo"`
	RetrievalCachePeriodSeconds                   types.Int64  `tfsdk:"retrieval_cache_period_seconds"`
}

// BaseVirtualRepositoryAPIModel contains common fields for all virtual repository API models
type BaseVirtualRepositoryAPIModel struct {
	repository.BaseRepositoryAPIModel
	Repositories                                  []string `json:"repositories"`
	ArtifactoryRequestsCanRetrieveRemoteArtifacts bool     `json:"artifactoryRequestsCanRetrieveRemoteArtifacts"`
	DefaultDeploymentRepo                         string   `json:"defaultDeploymentRepo"`
	RetrievalCachePeriodSeconds                   int64    `json:"virtualRetrievalCachePeriodSeconds"`
}

// BaseVirtualSchemaAttributes returns the base schema attributes for all virtual repository data sources
func BaseVirtualSchemaAttributes() map[string]schema.Attribute {
	return lo.Assign(
		repository.BaseSchemaAttributes(),
		map[string]schema.Attribute{
			"repositories": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The effective list of actual repositories included in this virtual repository.",
				Computed:            true,
			},
			"artifactory_requests_can_retrieve_remote_artifacts": schema.BoolAttribute{
				MarkdownDescription: "Whether the virtual repository should search through remote repositories when looking for artifacts in local repositories, or through a third-party artifact repository. A valid repository must be specified by `default_deployment_repo` when set to `false`.",
				Computed:            true,
			},
			"default_deployment_repo": schema.StringAttribute{
				MarkdownDescription: "Default repository to deploy artifacts.",
				Computed:            true,
			},
			"retrieval_cache_period_seconds": schema.Int64Attribute{
				MarkdownDescription: "The metadataRetrievalCachePeriod (in seconds) specifies how long the cache metadata should be considered valid.",
				Computed:            true,
			},
		},
	)
}

// CommonVirtualFromAPIModel provides common conversion logic from API model to Terraform model for virtual repositories
func CommonVirtualFromAPIModel(ctx context.Context, baseModel *BaseVirtualRepositoryDataSourceModel, apiModel BaseVirtualRepositoryAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// Convert common fields using the base repository function
	diags.Append(repository.CommonFromAPIModel(ctx, &baseModel.BaseRepositoryDataSourceModel, apiModel.BaseRepositoryAPIModel)...)
	if diags.HasError() {
		return diags
	}

	// Convert virtual-specific fields
	baseModel.ArtifactoryRequestsCanRetrieveRemoteArtifacts = types.BoolValue(apiModel.ArtifactoryRequestsCanRetrieveRemoteArtifacts)
	baseModel.DefaultDeploymentRepo = types.StringValue(apiModel.DefaultDeploymentRepo)
	baseModel.RetrievalCachePeriodSeconds = types.Int64Value(apiModel.RetrievalCachePeriodSeconds)

	// Convert repositories
	if apiModel.Repositories != nil {
		repositories, diags := types.SetValueFrom(ctx, types.StringType, apiModel.Repositories)
		if diags.HasError() {
			return diags
		}
		baseModel.Repositories = repositories
	} else {
		baseModel.Repositories = types.SetNull(types.StringType)
	}

	return diags
}
