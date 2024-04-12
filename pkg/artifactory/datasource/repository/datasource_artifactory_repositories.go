package repository

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
)

const EndPoint = "artifactory/api/repositories/"

var _ datasource.DataSource = &RepositoriesDataSource{}

func NewRepositoriesDataSource() datasource.DataSource {
	return &RepositoriesDataSource{}
}

type RepositoriesDataSource struct {
	ProviderData util.ProviderMetadata
}

type RepositoriesDataSourceModel struct {
	RepositoryType types.String `tfsdk:"repository_type"`
	PackageType    types.String `tfsdk:"package_type"`
	ProjectKey     types.String `tfsdk:"project_key"`
	Repos          types.Set    `tfsdk:"repos"`
}

type RepositoriesAPIModel struct {
	Key         string `json:"key"`
	Type        string `json:"type"`
	Description string `json:"description"`
	URL         string `json:"url"`
	PackageType string `json:"packageType"`
}

var reposAttrType = map[string]attr.Type{
	"key":          types.StringType,
	"type":         types.StringType,
	"description":  types.StringType,
	"url":          types.StringType,
	"package_type": types.StringType,
}

func (m *RepositoriesDataSourceModel) FromAPIModel(ctx context.Context, data []RepositoriesAPIModel) diag.Diagnostics {

	var repos []attr.Value

	for _, d := range data {
		repo := types.ObjectValueMust(
			reposAttrType,
			map[string]attr.Value{
				"key":          types.StringValue(d.Key),
				"type":         types.StringValue(d.Type),
				"description":  types.StringValue(d.Description),
				"url":          types.StringValue(d.URL),
				"package_type": types.StringValue(d.PackageType),
			},
		)

		repos = append(repos, repo)
	}

	reposSet, d := types.SetValue(types.ObjectType{AttrTypes: reposAttrType}, repos)
	if d != nil {
		return d
	}

	m.Repos = reposSet

	return nil
}

func (d *RepositoriesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "artifactory_repositories"
}

func (d *RepositoriesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"repository_type": schema.StringAttribute{
				Description: fmt.Sprintf("Filter for repositories of a specific type. Allowed values are: %s", strings.Join(validRepositoryTypes, ", ")),
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(validRepositoryTypes...),
				},
			},
			"package_type": schema.StringAttribute{
				Description: fmt.Sprintf("Filter for repositories of a specific package type. Allowed values are: %s", strings.Join(validPackageTypes, ", ")),
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(validPackageTypes...),
				},
			},
			"project_key": schema.StringAttribute{
				Description: "Filter for repositories assigned to a specific project.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z][a-z0-9\-]{1,31}$`),
						"project_key must be 2 - 32 lowercase alphanumeric and hyphen characters",
					),
				},
			},
			"repos": schema.SetNestedAttribute{
				Description: "A list of repositories.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key":          schema.StringAttribute{Computed: true},
						"type":         schema.StringAttribute{Computed: true},
						"description":  schema.StringAttribute{Computed: true},
						"url":          schema.StringAttribute{Computed: true},
						"package_type": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
		Description: "Returns a list of minimal repository details for all repositories of the specified type.",
	}
}

func (d *RepositoriesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	d.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (d *RepositoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RepositoriesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var repos []RepositoriesAPIModel
	response, err := d.ProviderData.Client.R().
		SetQueryParams(map[string]string{
			"type":        data.RepositoryType.ValueString(),
			"packageType": data.PackageType.ValueString(),
			"project":     data.ProjectKey.ValueString(),
		}).
		SetResult(&repos).
		Get(EndPoint)

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Data Source",
			"An unexpected error occurred while fetch the data source. "+
				"Please report this issue to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	if response.IsError() {
		resp.Diagnostics.AddError(
			"Unable to Read Data Source",
			"An unexpected error occurred while fetch the data source. "+
				"Please report this issue to the provider developers.\n\n"+
				"Error: "+response.String(),
		)
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(data.FromAPIModel(ctx, repos)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
