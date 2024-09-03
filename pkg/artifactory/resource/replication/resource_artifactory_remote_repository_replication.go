package replication

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"

	"github.com/jfrog/terraform-provider-shared/client"
)

var _ resource.Resource = &RemoteRepositoryReplicationResource{}

func NewRemoteRepositoryReplicationResource() resource.Resource {
	return &RemoteRepositoryReplicationResource{
		TypeName: "artifactory_remote_repository_replication",
	}
}

type RemoteRepositoryReplicationResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

type RemoteRepositoryReplicationResourceModel struct {
	ID                              types.String `tfsdk:"id"`
	EnableEventReplication          types.Bool   `tfsdk:"enable_event_replication"`
	Enabled                         types.Bool   `tfsdk:"enabled"`
	CronExp                         types.String `tfsdk:"cron_exp"`
	SyncDeletes                     types.Bool   `tfsdk:"sync_deletes"`
	SyncProperties                  types.Bool   `tfsdk:"sync_properties"`
	RepoKey                         types.String `tfsdk:"repo_key"`
	ReplicationKey                  types.String `tfsdk:"replication_key"`
	IncludePathPrefixPattern        types.String `tfsdk:"include_path_prefix_pattern"`
	ExcludePathPrefixPattern        types.String `tfsdk:"exclude_path_prefix_pattern"`
	CheckBinaryExistenceInFilestore types.Bool   `tfsdk:"check_binary_existence_in_filestore"`
}

func (m RemoteRepositoryReplicationResourceModel) toAPIModel(_ context.Context, apiModel *RemoteReplicationAPIModel) (diags diag.Diagnostics) {
	*apiModel = RemoteReplicationAPIModel{
		EnableEventReplication:          m.EnableEventReplication.ValueBool(),
		Enabled:                         m.Enabled.ValueBool(),
		CronExp:                         m.CronExp.ValueString(),
		SyncDeletes:                     m.SyncDeletes.ValueBool(),
		SyncProperties:                  m.SyncProperties.ValueBool(),
		RepoKey:                         m.RepoKey.ValueString(),
		IncludePathPrefixPattern:        m.IncludePathPrefixPattern.ValueString(),
		ExcludePathPrefixPattern:        m.ExcludePathPrefixPattern.ValueString(),
		CheckBinaryExistenceInFilestore: m.CheckBinaryExistenceInFilestore.ValueBool(),
	}

	return
}

func (m *RemoteRepositoryReplicationResourceModel) fromAPIModel(_ context.Context, apiModel RemoteReplicationGetAPIModel) (diags diag.Diagnostics) {
	m.ID = types.StringValue(apiModel.RepoKey)
	m.EnableEventReplication = types.BoolValue(apiModel.EnableEventReplication)
	m.Enabled = types.BoolValue(apiModel.Enabled)
	m.CronExp = types.StringValue(apiModel.CronExp)
	m.SyncDeletes = types.BoolValue(apiModel.SyncDeletes)
	m.SyncProperties = types.BoolValue(apiModel.SyncProperties)
	m.RepoKey = types.StringValue(apiModel.RepoKey)
	m.IncludePathPrefixPattern = types.StringValue(apiModel.IncludePathPrefixPattern)
	m.ExcludePathPrefixPattern = types.StringValue(apiModel.ExcludePathPrefixPattern)
	m.CheckBinaryExistenceInFilestore = types.BoolValue(apiModel.CheckBinaryExistenceInFilestore)
	m.ReplicationKey = types.StringValue(apiModel.ReplicationKey)

	return
}

type RemoteReplicationAPIModel struct {
	Enabled                         bool   `json:"enabled"`
	CronExp                         string `json:"cronExp"`
	SyncDeletes                     bool   `json:"syncDeletes"`
	SyncProperties                  bool   `json:"syncProperties"`
	IncludePathPrefixPattern        string `json:"includePathPrefixPattern"`
	ExcludePathPrefixPattern        string `json:"excludePathPrefixPattern"`
	RepoKey                         string `json:"repoKey"`
	ReplicationKey                  string `json:"replicationKey"`
	EnableEventReplication          bool   `json:"enableEventReplication"`
	CheckBinaryExistenceInFilestore bool   `json:"checkBinaryExistenceInFilestore"`
}

type RemoteReplicationGetAPIModel struct {
	RemoteReplicationAPIModel
	ReplicationKey string `json:"replicationKey"`
}

func (r *RemoteRepositoryReplicationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *RemoteRepositoryReplicationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"repo_key": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "Repository name.",
			},
			"cron_exp": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				MarkdownDescription: "The Cron expression that determines when the next replication will be triggered.",
			},
			"enable_event_replication": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When set, each event will trigger replication of the artifacts changed in this event. This can be any type of event on artifact, e.g. add, deleted or property change. Default value is `false`.",
			},
			"sync_deletes": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "When set, items that were deleted locally should also be deleted remotely (also applies to properties metadata). Note that enabling this option, will delete artifacts on the target that do not exist in the source repository. Default value is `false`.",
			},
			"sync_properties": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "When set, the task also synchronizes the properties of replicated artifacts. Default value is `true`.",
			},
			"enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "When set, enables replication of this repository to the target specified in `url` attribute. Default value is `true`.",
			},
			"include_path_prefix_pattern": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "List of artifact patterns to include when evaluating artifact requests in the form of x/y/**/z/*. When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included (**/*).",
			},
			"exclude_path_prefix_pattern": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "List of artifact patterns to exclude when evaluating artifact requests, in the form of x/y/**/z/*. By default no artifacts are excluded.",
			},
			"replication_key": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Replication ID.",
			},
			"check_binary_existence_in_filestore": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "Enabling the `check_binary_existence_in_filestore` flag requires an Enterprise+ license. When true, enables distributed checksum storage. For more information, see " +
					"[Optimizing Repository Replication with Checksum-Based Storage](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-OptimizingRepositoryReplicationUsingStorageLevelSynchronizationOptions).",
			},
		},
		MarkdownDescription: "Provides a local repository replication resource, also referred to as Artifactory push replication. This can be used to create and manage Artifactory local repository replications using [Push Replication API](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-SetRepositoryReplicationConfiguration).\n\n" +
			"Push replication is used to synchronize Local Repositories, and is implemented by the Artifactory server on the near end invoking a synchronization of artifacts to the far end. See the [Official Documentation](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-PushReplication).\n\n" +
			"This resource can create the replication of local repository to single repository on the remote server.",
	}
}

func (r *RemoteRepositoryReplicationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *RemoteRepositoryReplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan RemoteRepositoryReplicationResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if verified, err := verifyRepoRclass(plan.RepoKey.ValueString(), "remote", r.ProviderData.Client.R()); !verified {
		utilfw.UnableToCreateResourceError(resp, fmt.Sprintf("source repository rclass is not remote, only remote repositories are supported by this resource %v", err))
		return
	}

	var replication RemoteReplicationAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &replication)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.ProviderData.Client.R().
		SetPathParam("repo_key", plan.RepoKey.ValueString()).
		SetBody(replication).
		Put(ReplicationEndpoint)

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	plan.ID = plan.RepoKey
	plan.ReplicationKey = types.StringNull()

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RemoteRepositoryReplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state RemoteRepositoryReplicationResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var replication RemoteReplicationGetAPIModel
	response, err := r.ProviderData.Client.R().
		SetPathParam("repo_key", state.RepoKey.ValueString()).
		SetResult(&replication).
		Get(ReplicationEndpoint)

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

	resp.Diagnostics.Append(state.fromAPIModel(ctx, replication)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RemoteRepositoryReplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan RemoteRepositoryReplicationResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state RemoteRepositoryReplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if verified, err := verifyRepoRclass(plan.RepoKey.ValueString(), "remote", r.ProviderData.Client.R()); !verified {
		utilfw.UnableToUpdateResourceError(resp, fmt.Sprintf("source repository rclass is not remote, only remote repositories are supported by this resource %v", err))
		return
	}

	var replication RemoteReplicationAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &replication)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.ProviderData.Client.R().
		SetPathParam("repo_key", plan.RepoKey.ValueString()).
		SetBody(replication).
		AddRetryCondition(client.RetryOnMergeError).
		Post(ReplicationEndpoint)

	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToUpdateResourceError(resp, response.String())
		return
	}

	plan.ID = state.ID
	plan.ReplicationKey = state.ReplicationKey

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RemoteRepositoryReplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state RemoteRepositoryReplicationResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	response, err := r.ProviderData.Client.R().
		SetPathParam("repo_key", state.RepoKey.ValueString()).
		AddRetryCondition(client.RetryOnMergeError).
		Delete(ReplicationEndpoint)

	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	if response.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if response.IsError() {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *RemoteRepositoryReplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("repo_key"), req, resp)
}
