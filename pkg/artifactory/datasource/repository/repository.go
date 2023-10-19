package repository

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	diagsdk "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	schemasdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"

	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
)

type DataSourceModel struct {
	Key                 types.String `tfsdk:"key"`
	ProjectKey          types.String `tfsdk:"project_key"`
	ProjectEnvironments types.Set    `tfsdk:"project_environments"`
	PackageType         types.String `tfsdk:"package_type"`
	Description         types.String `tfsdk:"description"`
	Notes               types.String `tfsdk:"notes"`
	IncludesPattern     types.Set    `tfsdk:"includes_pattern"`
	ExcludesPattern     types.Set    `tfsdk:"excludes_pattern"`
	RepoLayoutRef       types.String `tfsdk:"repo_layout_ref"`
}

func (m DataSourceModel) FromAPIModel(ctx context.Context, data APIModel) diag.Diagnostics {
	return nil
}

var RepoSchema map[string]schema.Attribute = map[string]schema.Attribute{
	"key":              schema.StringAttribute{Computed: true},
	"project_key":      schema.StringAttribute{Computed: true},
	"environments":     schema.SetAttribute{ElementType: types.StringType, Computed: true},
	"package_type":     schema.StringAttribute{Computed: true},
	"description":      schema.StringAttribute{Computed: true},
	"notes":            schema.StringAttribute{Computed: true},
	"includes_pattern": schema.StringAttribute{Computed: true},
	"excludes_pattern": schema.StringAttribute{Computed: true},
	"repo_layout_ref":  schema.StringAttribute{Computed: true},
}

type APIModel struct {
	Key             string   `json:"key"`
	ProjectKey      string   `json:"project_key"`
	Environments    []string `json:"environments"`
	PackageType     string   `json:"package_type"`
	Description     string   `json:"description"`
	Notes           string   `json:"notes"`
	IncludesPattern string   `json:"includes_patterns"`
	ExcludesPattern string   `json:"excludes_patterns"`
}

const EndPoint = "artifactory/api/repositories/"

func MkRepoReadDataSource(pack packer.PackFunc, construct repository.Constructor) schemasdk.ReadContextFunc {
	return func(ctx context.Context, d *schemasdk.ResourceData, m interface{}) diagsdk.Diagnostics {
		repo, err := construct()
		if err != nil {
			return diagsdk.FromErr(err)
		}

		key := d.Get("key").(string)
		// repo must be a pointer
		_, err = m.(utilsdk.ProvderMetadata).Client.R().
			SetResult(repo).
			SetPathParam("key", key).
			Get(repository.RepositoriesEndpoint)

		if err != nil {
			return diagsdk.FromErr(err)
		}

		d.SetId(key)

		return diagsdk.FromErr(pack(repo, d))
	}
}
