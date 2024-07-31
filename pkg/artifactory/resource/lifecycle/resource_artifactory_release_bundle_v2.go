package lifecycle

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"
)

const (
	ReleaseBundleV2Endpoint        = "lifecycle/api/v2/release_bundle"
	ReleaseBundleV2VersionEndpoint = "lifecycle/api/v2/release_bundle/records/{name}/{version}"
)

var _ resource.Resource = &ReleaseBundleV2Resource{}

func NewReleaseBundleV2Resource() resource.Resource {
	return &ReleaseBundleV2Resource{
		TypeName: "artifactory_release_bundle_v2",
	}
}

type ReleaseBundleV2Resource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

type ReleaseBundleV2ResourceModel struct {
	Name                         types.String `tfsdk:"name"`
	Version                      types.String `tfsdk:"version"`
	KeyPairName                  types.String `tfsdk:"keypair_name"`
	ProjectKey                   types.String `tfsdk:"project_key"`
	SkipDockerManifestResolution types.Bool   `tfsdk:"skip_docker_manifest_resolution"`
	SourceType                   types.String `tfsdk:"source_type"`
	Source                       types.Object `tfsdk:"source"`
	Created                      types.String `tfsdk:"created"`
	CreatedBy                    types.String `tfsdk:"created_by"`
	ServiceID                    types.String `tfsdk:"service_id"`
}

func (m ReleaseBundleV2ResourceModel) toAPIModel(_ context.Context, apiModel *ReleaseBundleV2RequestAPIModel) (diags diag.Diagnostics) {
	sourceType := m.SourceType.ValueString()
	source := ReleaseBundleV2SourceAPIModel{
		AQL:            "",
		Artifacts:      []ReleaseBundleV2SourceArtifactAPIModel{},
		Builds:         []ReleaseBundleV2SourceBuildAPIModel{},
		ReleaseBundles: []ReleaseBundleV2SourceReleaseBundleAPIModel{},
	}

	sourceAttrs := m.Source.Attributes()

	switch sourceType {
	case "aql":
		aql, ok := sourceAttrs[sourceType]
		if !ok {
			diags.AddAttributeError(
				path.Root("source").AtName(sourceType),
				"failed to access source attribute value",
				"",
			)
		}

		source.AQL = aql.(types.String).ValueString()

	case "artifacts":
		artifactsSet, ok := sourceAttrs[sourceType]
		if !ok {
			diags.AddAttributeError(
				path.Root("source").AtName(sourceType),
				"failed to access source attribute value",
				"",
			)
		}

		artifacts := lo.Map(
			artifactsSet.(types.Set).Elements(),
			func(elem attr.Value, _ int) ReleaseBundleV2SourceArtifactAPIModel {
				attrs := elem.(types.Object).Attributes()

				return ReleaseBundleV2SourceArtifactAPIModel{
					Path:   attrs["path"].(types.String).ValueString(),
					SHA256: attrs["sha256"].(types.String).ValueString(),
				}
			},
		)

		source.Artifacts = artifacts

	case "builds":
		buildsSet, ok := sourceAttrs[sourceType]
		if !ok {
			diags.AddAttributeError(
				path.Root("source").AtName(sourceType),
				"failed to access source attribute value",
				"",
			)
		}

		builds := lo.Map(
			buildsSet.(types.Set).Elements(),
			func(elem attr.Value, _ int) ReleaseBundleV2SourceBuildAPIModel {
				attrs := elem.(types.Object).Attributes()

				return ReleaseBundleV2SourceBuildAPIModel{
					Repository:          attrs["repository"].(types.String).ValueString(),
					Name:                attrs["name"].(types.String).ValueString(),
					Number:              attrs["number"].(types.String).ValueString(),
					Started:             attrs["started"].(types.String).ValueString(),
					IncludeDependencies: attrs["include_dependencies"].(types.Bool).ValueBool(),
				}
			},
		)

		source.Builds = builds

	case "release_bundles":
		releaseBundlesSet, ok := sourceAttrs[sourceType]
		if !ok {
			diags.AddAttributeError(
				path.Root("source").AtName(sourceType),
				"failed to access source attribute value",
				"",
			)
		}

		releaseBundles := lo.Map(
			releaseBundlesSet.(types.Set).Elements(),
			func(elem attr.Value, _ int) ReleaseBundleV2SourceReleaseBundleAPIModel {
				attrs := elem.(types.Object).Attributes()

				return ReleaseBundleV2SourceReleaseBundleAPIModel{
					ProjectKey:           attrs["project_key"].(types.String).ValueString(),
					RepositoryKey:        attrs["repository_key"].(types.String).ValueString(),
					ReleaseBundleName:    attrs["name"].(types.String).ValueString(),
					ReleaseBundleVersion: attrs["version"].(types.String).ValueString(),
				}
			},
		)

		source.ReleaseBundles = releaseBundles
	}

	*apiModel = ReleaseBundleV2RequestAPIModel{
		Name:                         m.Name.ValueString(),
		Version:                      m.Version.ValueString(),
		SkipDockerManifestResolution: m.SkipDockerManifestResolution.ValueBool(),
		SourceType:                   sourceType,
		Source:                       source,
	}

	return
}

type ReleaseBundleV2RequestAPIModel struct {
	Name                         string                        `json:"release_bundle_name"`
	Version                      string                        `json:"release_bundle_version"`
	SkipDockerManifestResolution bool                          `json:"skip_docker_manifest_resolution"`
	SourceType                   string                        `json:"source_type"`
	Source                       ReleaseBundleV2SourceAPIModel `json:"source"`
}

type ReleaseBundleV2SourceAPIModel struct {
	AQL            string                                       `json:"aql,omitempty"`
	Artifacts      []ReleaseBundleV2SourceArtifactAPIModel      `json:"artifacts,omitempty"`
	Builds         []ReleaseBundleV2SourceBuildAPIModel         `json:"builds,omitempty"`
	ReleaseBundles []ReleaseBundleV2SourceReleaseBundleAPIModel `json:"release_bundles,omitempty"`
}

type ReleaseBundleV2SourceArtifactAPIModel struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256,omitempty"`
}

type ReleaseBundleV2SourceBuildAPIModel struct {
	Repository          string `json:"build_repository,omitempty"`
	Name                string `json:"build_name"`
	Number              string `json:"build_number"`
	Started             string `json:"build_started,omitempty"`
	IncludeDependencies bool   `json:"include_dependencies"`
}

type ReleaseBundleV2SourceReleaseBundleAPIModel struct {
	ProjectKey           string `json:"project_key,omitempty"`
	RepositoryKey        string `json:"repository_key,omitempty"`
	ReleaseBundleName    string `json:"release_bundle_name"`
	ReleaseBundleVersion string `json:"release_bundle_version"`
}

type ReleaseBundleV2ResponseAPIModel struct {
	RepositoryKey string `json:"repository_key"`
	Name          string `json:"release_bundle_name"`
	Version       string `json:"release_bundle_version"`
	Created       string `json:"created"`
}

type ReleaseBundleV2GetAPIModel struct {
	ServiceID string `json:"service_id"`
	CreatedBy string `json:"created_by"`
	Created   string `json:"created"`
}

func (r *ReleaseBundleV2Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *ReleaseBundleV2Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexp.MustCompile(`[a-zA-Z_0-9][a-zA-Z_\.\-0-9]*`), "Must begin with [a-z A-Z _ 0-9] and consist of [a-z A-Z _ . - 0-9]"),
				},
				Description: "Name of Release Bundle",
			},
			"version": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
					stringvalidator.RegexMatches(regexp.MustCompile(`[a-zA-Z_0-9][a-zA-Z_\.\-0-9]*`), "Must begin with [a-z A-Z _ 0-9] and consist of [a-z A-Z _ . - 0-9]"),
				},
				Description: "Version to promote",
			},
			"keypair_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "Key-pair name to use for signature creation",
			},
			"project_key": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					validatorfw_string.ProjectKey(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Project key the Release Bundle belongs to",
			},
			"skip_docker_manifest_resolution": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Determines whether to skip the resolution of the Docker manifest, which adds the image layers to the Release Bundle. The default value is `false` (the manifest is resolved and image layers are included).",
			},
			"source_type": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.OneOf("aql", "artifacts", "builds", "release_bundles"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Source type. Valid values: `aql`, `artifacts`, `builds`, `release_bundles`",
			},
			"source": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"aql": schema.StringAttribute{
						Optional: true,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(
								path.MatchRelative().AtParent().AtName("artifacts"),
								path.MatchRelative().AtParent().AtName("builds"),
								path.MatchRelative().AtParent().AtName("release_bundles"),
							),
						},
						MarkdownDescription: "The contents of the AQL query.",
					},
					"artifacts": schema.SetNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"path": schema.StringAttribute{
									Required: true,
									Validators: []validator.String{
										stringvalidator.LengthAtLeast(1),
									},
									MarkdownDescription: "The path for the artifact",
								},
								"sha256": schema.StringAttribute{
									Optional: true,
									Validators: []validator.String{
										stringvalidator.LengthAtLeast(1),
									},
									MarkdownDescription: "The SHA256 for the artifact",
								},
							},
						},
						Optional: true,
						Validators: []validator.Set{
							setvalidator.ConflictsWith(
								path.MatchRelative().AtParent().AtName("aql"),
								path.MatchRelative().AtParent().AtName("builds"),
								path.MatchRelative().AtParent().AtName("release_bundles"),
							),
						},
						MarkdownDescription: "Source type to create a Release Bundle v2 version by collecting source artifacts from a list of path/checksum pairs.",
					},
					"builds": schema.SetNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required: true,
									Validators: []validator.String{
										stringvalidator.LengthAtLeast(1),
									},
									MarkdownDescription: "Name of the build.",
								},
								"number": schema.StringAttribute{
									Required: true,
									Validators: []validator.String{
										stringvalidator.LengthAtLeast(1),
									},
									MarkdownDescription: "Number (run) of the build.",
								},
								"started": schema.StringAttribute{
									Optional: true,
									Validators: []validator.String{
										stringvalidator.LengthAtLeast(1),
									},
									MarkdownDescription: "Timestamp when the build was created. If omitted, the system uses the latest build run, as identified by the `name` and `number` combination. The timestamp is provided according to the ISO 8601 standard.",
								},
								"repository": schema.StringAttribute{
									Optional: true,
									Validators: []validator.String{
										stringvalidator.LengthAtLeast(1),
									},
									MarkdownDescription: "The repository key of the build. If omitted, the system uses the default built-in repository, `artifactory-build-info`.",
								},
								"include_dependencies": schema.BoolAttribute{
									Optional: true,
									Computed: true,
									Default:  booldefault.StaticBool(false),
									MarkdownDescription: "Determines whether to include build dependencies in the Release Bundle. The default value is `false`.\n\n" +
										"~>Dependencies must be located in local or Federated repositories to be included in the Release Bundle. Dependencies located in remote repositories are not supported.",
								},
							},
						},
						Optional: true,
						Validators: []validator.Set{
							setvalidator.ConflictsWith(
								path.MatchRelative().AtParent().AtName("aql"),
								path.MatchRelative().AtParent().AtName("artifacts"),
								path.MatchRelative().AtParent().AtName("release_bundles"),
							),
						},
						MarkdownDescription: "Source type to create a Release Bundle v2 version by collecting source artifacts from one or multiple builds (also known as build-info).",
					},
					"release_bundles": schema.SetNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"project_key": schema.StringAttribute{
									Optional: true,
									Validators: []validator.String{
										stringvalidator.LengthAtLeast(1),
										validatorfw_string.ProjectKey(),
									},
									MarkdownDescription: "Project key of the release bundle.",
								},
								"repository_key": schema.StringAttribute{
									Optional: true,
									Validators: []validator.String{
										stringvalidator.LengthAtLeast(1),
									},
									MarkdownDescription: "The key of the release bundle repository.",
								},
								"name": schema.StringAttribute{
									Required: true,
									Validators: []validator.String{
										stringvalidator.LengthAtLeast(1),
									},
									MarkdownDescription: "The name of the release bundle.",
								},
								"version": schema.StringAttribute{
									Required: true,
									Validators: []validator.String{
										stringvalidator.LengthAtLeast(1),
									},
									MarkdownDescription: "The version of the release bundle.",
								},
							},
						},
						Optional: true,
						Validators: []validator.Set{
							setvalidator.ConflictsWith(
								path.MatchRelative().AtParent().AtName("aql"),
								path.MatchRelative().AtParent().AtName("artifacts"),
								path.MatchRelative().AtParent().AtName("builds"),
							),
						},
						MarkdownDescription: "Source type to create a Release Bundle v2 version by collecting source artifacts from existing Release Bundle versions. Must match `source_type` attribute value.",
					},
				},
				Required:            true,
				MarkdownDescription: "Defines specific repositories to include in the promotion. If this property is left undefined, all repositories (except those specifically excluded) are included in the promotion. Important: If one or more repositories are specifically included, all other repositories are excluded (regardless of what is defined in `excluded_repository_keys`).",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the new version was created (ISO 8601 standard).",
			},
			"created_by": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The user who created the Release Bundle.",
			},
			"service_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the Artifactory instance where the Release Bundle was created.",
			},
		},
		MarkdownDescription: "This resource enables you to creates a new Release Bundle v2, uniquely identified by a combination of repository key, name, and version. For more information, see [Understanding Release Bundles v2](https://jfrog.com/help/r/jfrog-artifactory-documentation/understanding-release-bundles-v2) and [REST API](https://jfrog.com/help/r/jfrog-rest-apis/create-release-bundle-v2-version).",
	}
}

func (r ReleaseBundleV2Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ReleaseBundleV2ResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sourceType := data.SourceType.ValueString()
	sourceAttrs := data.Source.Attributes()

	if _, ok := sourceAttrs[sourceType]; !ok {
		resp.Diagnostics.AddAttributeError(
			path.Root("source").AtName(sourceType),
			"Invalid Attribute Configuration",
			fmt.Sprintf("Expected source.%s to be configured with source_type is set to '%s'.", sourceType, sourceType),
		)
	}
}

func (r *ReleaseBundleV2Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *ReleaseBundleV2Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ReleaseBundleV2ResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var releaseBundle ReleaseBundleV2RequestAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &releaseBundle)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.ProviderData.Client.R().
		SetHeader("X-JFrog-Signing-Key-Name", plan.KeyPairName.ValueString()).
		SetQueryParam("async", "false")

	if !plan.ProjectKey.IsNull() {
		request.SetQueryParam("project", plan.ProjectKey.ValueString())
	}

	var result ReleaseBundleV2ResponseAPIModel

	response, err := request.
		SetBody(releaseBundle).
		SetResult(&result).
		Post(ReleaseBundleV2Endpoint)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	plan.Created = types.StringValue(result.Created)
	plan.CreatedBy = types.StringNull()
	plan.ServiceID = types.StringNull()

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ReleaseBundleV2Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ReleaseBundleV2ResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var releaseBundle ReleaseBundleV2GetAPIModel

	request := r.ProviderData.Client.R()

	if !state.ProjectKey.IsNull() {
		request.SetQueryParam("project", state.ProjectKey.ValueString())
	}

	response, err := request.
		SetPathParams(map[string]string{
			"name":    state.Name.ValueString(),
			"version": state.Version.ValueString(),
		}).
		SetResult(&releaseBundle).
		Get(ReleaseBundleV2VersionEndpoint)
	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	if response.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	state.Created = types.StringValue(releaseBundle.Created)
	state.CreatedBy = types.StringValue(releaseBundle.CreatedBy)
	state.ServiceID = types.StringValue(releaseBundle.ServiceID)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ReleaseBundleV2Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning(
		"Update not supported",
		"Release Bundle V2 cannnot be updated.",
	)
}

func (r *ReleaseBundleV2Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ReleaseBundleV2ResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	request := r.ProviderData.Client.R()

	if !state.ProjectKey.IsNull() {
		request.SetQueryParam("project", state.ProjectKey.ValueString())
	}

	response, err := request.
		SetPathParams(map[string]string{
			"name":    state.Name.ValueString(),
			"version": state.Version.ValueString(),
		}).
		SetQueryParam("async", "false").
		Delete(ReleaseBundleV2VersionEndpoint)

	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}
