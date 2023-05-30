package security

import (
	"context"
	"net/http"

	// "github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	validatorfw "github.com/jfrog/terraform-provider-shared/validator/fw"
)

const PermissionsEndPoint = "artifactory/api/v2/security/permissions/"

func NewPermissionTargetResource() resource.Resource {
	return &PermissionTargetResource{}
}

type PermissionTargetResource struct {
	ProviderData utilsdk.ProvderMetadata
}

// PermissionTargetResourceModel describes the Terraform resource data model to match the
// resource schema.
type PermissionTargetResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Repo          types.Object `tfsdk:"repo"`
	Build         types.Object `tfsdk:"build"`
	ReleaseBundle types.Object `tfsdk:"release_bundle"`
}

func (r *PermissionTargetResourceModel) toActionsAPIModel(ctx context.Context, resourceActions types.Object) Actions {
	setToMap := func(sourceSet types.Set, mapToUpdate *map[string][]string) {
		setElements := sourceSet.Elements()
		for _, setElement := range setElements {
			attrs := setElement.(types.Object).Attributes()
			permissions := utilfw.StringSetToStrings(attrs["permissions"].(types.Set))
			(*mapToUpdate)[attrs["name"].(types.String).ValueString()] = permissions
		}
	}

	actions := Actions{
		Users:  map[string][]string{},
		Groups: map[string][]string{},
	}
	actionsAttrs := resourceActions.Attributes()

	setToMap(actionsAttrs["users"].(types.Set), &actions.Users)
	setToMap(actionsAttrs["groups"].(types.Set), &actions.Groups)

	return actions
}

func (r *PermissionTargetResourceModel) toSectionAPIModel(ctx context.Context, resourceSection types.Object) *PermissionTargetSection {
	tflog.Debug(ctx, "toSectionAPIModel", map[string]interface{}{
		"resourceSection": resourceSection,
	})
	if resourceSection.IsUnknown() || resourceSection.IsNull() {
		return nil
	}
	sectionAttrs := resourceSection.Attributes()
	repoActions := r.toActionsAPIModel(ctx, sectionAttrs["actions"].(types.Object))

	return &PermissionTargetSection{
		IncludePatterns: utilfw.StringSetToStrings(sectionAttrs["includes_pattern"].(types.Set)),
		ExcludePatterns: utilfw.StringSetToStrings(sectionAttrs["excludes_pattern"].(types.Set)),
		Repositories:    utilfw.StringSetToStrings(sectionAttrs["repositories"].(types.Set)),
		Actions:         &repoActions,
	}
}

func (r *PermissionTargetResourceModel) toAPIModel(ctx context.Context) PermissionTargetResourceAPIModel {
	// convert section
	repo := r.toSectionAPIModel(ctx, r.Repo)
	build := r.toSectionAPIModel(ctx, r.Build)
	releaseBundle := r.toSectionAPIModel(ctx, r.ReleaseBundle)

	// Convert from Terraform data model into API data model
	return PermissionTargetResourceAPIModel{
		Name:          r.Name.ValueString(),
		Repo:          repo,
		Build:         build,
		ReleaseBundle: releaseBundle,
	}
}

func (r *PermissionTargetResourceModel) ToState(ctx context.Context, diags diag.Diagnostics, permissionTarget *PermissionTargetResourceAPIModel) {
	r.Id = types.StringValue(permissionTarget.Name)
	r.Name = types.StringValue(permissionTarget.Name)

	var ds diag.Diagnostics

	if permissionTarget.Repo != nil {
		r.Repo, ds = r.sectionFromAPIModel(ctx, permissionTarget.Repo)
		if ds != nil {
			diags.Append(ds...)
			return
		}
	}

	if permissionTarget.Build != nil {
		r.Build, ds = r.sectionFromAPIModel(ctx, permissionTarget.Build)
		if ds != nil {
			diags.Append(ds...)
			return
		}
	}

	if permissionTarget.ReleaseBundle != nil {
		r.ReleaseBundle, ds = r.sectionFromAPIModel(ctx, permissionTarget.ReleaseBundle)
		if ds != nil {
			diags.Append(ds...)
			return
		}
	}
}

var namePermissionAttrTypes = map[string]attr.Type{
	"name":        types.StringType,
	"permissions": types.SetType{types.StringType},
}

var actionsAttrTypes = map[string]attr.Type{
	"users": types.SetType{
		types.ObjectType{
			AttrTypes: namePermissionAttrTypes,
		},
	},
	"groups": types.SetType{
		types.ObjectType{
			AttrTypes: namePermissionAttrTypes,
		},
	},
}

func (r *PermissionTargetResourceModel) sectionFromAPIModel(ctx context.Context, section *PermissionTargetSection) (types.Object, diag.Diagnostics) {
	includesPatterns, diags := types.SetValueFrom(ctx, types.StringType, section.IncludePatterns)
	if diags != nil {
		return types.Object{}, diags
	}

	excludesPatterns, diags := types.SetValueFrom(ctx, types.StringType, section.ExcludePatterns)
	if diags != nil {
		return types.Object{}, diags
	}

	repos, diags := types.SetValueFrom(ctx, types.StringType, section.Repositories)
	if diags != nil {
		return types.Object{}, diags
	}

	actionFromAPIModel := func(action map[string][]string) (types.Set, diag.Diagnostics) {
		objectValues := []attr.Value{}
		for name, permissions := range action {
			permissionsSet, diags := types.SetValueFrom(ctx, types.StringType, permissions)
			if diags != nil {
				return types.Set{}, diags
			}

			objectValue := types.ObjectValueMust(
				namePermissionAttrTypes,
				map[string]attr.Value{
					"name":        types.StringValue(name),
					"permissions": permissionsSet,
				},
			)
			objectValues = append(objectValues, objectValue)
		}
		return types.SetValueMust(types.ObjectType{
			AttrTypes: namePermissionAttrTypes,
		}, objectValues), nil
	}

	actionsUsers, diags := actionFromAPIModel(section.Actions.Users)
	if diags != nil {
		return types.Object{}, diags
	}

	actionsGroups, diags := actionFromAPIModel(section.Actions.Groups)
	if diags != nil {
		return types.Object{}, diags
	}

	return types.ObjectValueMust(
		map[string]attr.Type{
			"includes_pattern": types.SetType{types.StringType},
			"excludes_pattern": types.SetType{types.StringType},
			"repositories":     types.SetType{types.StringType},
			"actions": types.ObjectType{
				AttrTypes: actionsAttrTypes,
			},
		}, map[string]attr.Value{
			"includes_pattern": includesPatterns,
			"excludes_pattern": excludesPatterns,
			"repositories":     repos,
			"actions": types.ObjectValueMust(
				actionsAttrTypes,
				map[string]attr.Value{
					"users":  actionsUsers,
					"groups": actionsGroups,
				},
			),
		},
	), nil
}

// PermissionTargetResourceAPIModel describes the API data model. Copy from https://github.com/jfrog/jfrog-client-go/blob/master/artifactory/services/permissiontarget.go#L116
//
// Using struct pointers to keep the fields null if they are empty.
// Artifactory evaluates inner struct typed fields if they are not null, which can lead to failures in the request.
type PermissionTargetResourceAPIModel struct {
	Name          string                   `json:"name"`
	Repo          *PermissionTargetSection `json:"repo,omitempty"`
	Build         *PermissionTargetSection `json:"build,omitempty"`
	ReleaseBundle *PermissionTargetSection `json:"releaseBundle,omitempty"`
}

type PermissionTargetSection struct {
	IncludePatterns []string `json:"include-patterns,omitempty"`
	ExcludePatterns []string `json:"exclude-patterns,omitempty"`
	Repositories    []string `json:"repositories"`
	Actions         *Actions `json:"actions,omitempty"`
}

type Actions struct {
	Users  map[string][]string `json:"users,omitempty"`
	Groups map[string][]string `json:"groups,omitempty"`
}

const (
	PermRead            = "read"
	PermWrite           = "write"
	PermAnnotate        = "annotate"
	PermDelete          = "delete"
	PermManage          = "manage"
	PermManagedXrayMeta = "managedXrayMeta"
	PermDistribute      = "distribute"
)

func (r *PermissionTargetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "artifactory_permission_target"
}

var actionsAttributeBlock = schema.SetNestedBlock{
	NestedObject: schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"permissions": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					validatorfw.StringInSlice([]string{
						PermRead,
						PermAnnotate,
						PermWrite,
						PermDelete,
						PermManage,
						PermManagedXrayMeta,
						PermDistribute,
					}),
				},
			},
		},
	},
}

func (r *PermissionTargetResource) getPrincipalBlock(description, repoDescription string) schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Attributes: map[string]schema.Attribute{
			"includes_pattern": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(basetypes.NewSetValueMust(types.StringType, []attr.Value{types.StringValue("**")})),
				MarkdownDescription: `The default value will be [""] if nothing is supplied`,
			},
			"excludes_pattern": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             setdefault.StaticValue(basetypes.NewSetValueMust(types.StringType, []attr.Value{})),
				MarkdownDescription: "The default value will be [] if nothing is supplied",
			},
			"repositories": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: repoDescription,
			},
		},
		Validators: []validator.Object{
			validatorfw.RequireIfDefined(path.Expressions{
				path.MatchRelative().AtName("repositories"),
			}...),
		},
		Blocks: map[string]schema.Block{
			"actions": schema.SingleNestedBlock{
				Blocks: map[string]schema.Block{
					"users":  actionsAttributeBlock,
					"groups": actionsAttributeBlock,
				},
			},
		},
		MarkdownDescription: description,
	}
}

func (r *PermissionTargetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides an Artifactory permission target resource. This can be used to create and manage Artifactory permission targets.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of permission.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"repo":           r.getPrincipalBlock("Repository permission configuration.", "You can specify the name `ANY` in the repositories section in order to apply to all repositories, `ANY REMOTE` for all remote repositories and `ANY LOCAL` for all local repositories. The default value will be [] if nothing is specified."),
			"build":          r.getPrincipalBlock("As for repo but for artifactory-build-info permissions.", `This can only be 1 value: "artifactory-build-info", and currently, validation of sets/lists is not allowed. Artifactory will reject the request if you change this`),
			"release_bundle": r.getPrincipalBlock("As for repo for for release-bundles permissions.", "You can specify the name `ANY` in the repositories section in order to apply to all repositories, `ANY REMOTE` for all remote repositories and `ANY LOCAL` for all local repositories. The default value will be [] if nothing is specified."),
		},
	}
}

func (r *PermissionTargetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(utilsdk.ProvderMetadata)
}

func (r *PermissionTargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *PermissionTargetResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	permissionTarget := data.toAPIModel(ctx)

	response, err := r.ProviderData.Client.R().
		SetBody(permissionTarget).
		Put(PermissionsEndPoint + permissionTarget.Name)

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode() != http.StatusOK {
		utilfw.UnableToCreateResourceError(resp, response.Status())
		return
	}

	// Assign the resource ID for the resource in the state
	data.Id = types.StringValue(permissionTarget.Name)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionTargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *PermissionTargetResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	permissionTarget := &PermissionTargetResourceAPIModel{}

	response, err := r.ProviderData.Client.R().
		SetResult(permissionTarget).
		Get(PermissionsEndPoint + data.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Refresh Resource",
			"An unexpected error occurred while attempting to refresh resource state. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Error: "+err.Error(),
		)

		return
	}

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if response.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)

		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	data.ToState(ctx, resp.Diagnostics, permissionTarget)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionTargetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *PermissionTargetResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// Convert from Terraform data model into API data model
	permissionTarget := data.toAPIModel(ctx)

	// Update call
	response, err := r.ProviderData.Client.R().
		SetBody(permissionTarget).
		Put(PermissionsEndPoint + permissionTarget.Name)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode() != http.StatusOK {
		utilfw.UnableToUpdateResourceError(resp, response.Status())
		return
	}

	data.ToState(ctx, resp.Diagnostics, &permissionTarget)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PermissionTargetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PermissionTargetResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	response, err := r.ProviderData.Client.R().
		Delete(PermissionsEndPoint + data.Id.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while attempting to delete the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Error: "+err.Error(),
		)

		return
	}

	// Return error if the HTTP status code is not 200 OK or 404 Not Found
	if response.StatusCode() != http.StatusNotFound && response.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			"An unexpected error occurred while attempting to delete the resource. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"HTTP Status: "+response.Status(),
		)

		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *PermissionTargetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// PermissionTargetParams Copy from https://github.com/jfrog/jfrog-client-go/blob/master/artifactory/services/permissiontarget.go#L116
//
// Using struct pointers to keep the fields null if they are empty.
// Artifactory evaluates inner struct typed fields if they are not null, which can lead to failures in the request.

func PermTargetExists(id string, m interface{}) (bool, error) {
	resp, err := m.(utilsdk.ProvderMetadata).Client.R().Head(PermissionsEndPoint + id)
	if err != nil && resp != nil && resp.StatusCode() == http.StatusNotFound {
		// Do not error on 404s as this causes errors when the upstream permission has been manually removed
		return false, nil
	}

	return err == nil, err
}
