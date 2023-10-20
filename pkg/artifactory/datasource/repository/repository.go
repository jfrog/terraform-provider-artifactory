package repository

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	Key             types.String `tfsdk:"key"`
	ProjectKey      types.String `tfsdk:"project_key"`
	Environments    types.Set    `tfsdk:"environments"`
	PackageType     types.String `tfsdk:"package_type"`
	Description     types.String `tfsdk:"description"`
	Notes           types.String `tfsdk:"notes"`
	IncludesPattern types.String `tfsdk:"includes_pattern"`
	ExcludesPattern types.String `tfsdk:"excludes_pattern"`
	RepoLayoutRef   types.String `tfsdk:"repo_layout_ref"`
}

func (m DataSourceModel) SetValueFromAPIModel(ctx context.Context, data APIModel) (map[string]attr.Value, diag.Diagnostics) {
	environments, diag := types.SetValueFrom(ctx, types.StringType, data.Environments)
	if diag != nil {
		return nil, diag
	}

	value := map[string]attr.Value{
		"key":              types.StringValue(data.Key),
		"project_key":      types.StringValue(data.ProjectKey),
		"environments":     environments,
		"package_type":     types.StringValue(data.PackageType),
		"description":      types.StringValue(data.Description),
		"notes":            types.StringValue(data.Notes),
		"includes_pattern": types.StringValue(data.IncludesPattern),
		"excludes_pattern": types.StringValue(data.ExcludesPattern),
		"repo_layout_ref":  types.StringValue(data.RepoLayoutRef),
	}
	return value, nil
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
	IncludesPattern string   `json:"includes_pattern"`
	ExcludesPattern string   `json:"excludes_pattern"`
	RepoLayoutRef   string   `json:"repo_layout_ref"`
}

var BaseAttrType = map[string]attr.Type{
	"key":              types.StringType,
	"project_key":      types.StringType,
	"environments":     types.SetType{ElemType: types.StringType},
	"package_type":     types.StringType,
	"description":      types.StringType,
	"notes":            types.StringType,
	"includes_pattern": types.StringType,
	"excludes_pattern": types.StringType,
	"repo_layout_ref":  types.StringType,
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
