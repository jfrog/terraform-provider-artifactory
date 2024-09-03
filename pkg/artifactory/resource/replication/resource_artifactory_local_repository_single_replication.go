package replication

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"

	"github.com/jfrog/terraform-provider-shared/client"
)

var _ resource.Resource = &LocalRepositorySingleReplicationResource{}

func NewLocalRepositorySingleReplicationResource() resource.Resource {
	return &LocalRepositorySingleReplicationResource{
		TypeName: "artifactory_local_repository_single_replication",
	}
}

type LocalRepositorySingleReplicationResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

type LocalRepositorySingleReplicationResourceModel struct {
	ID                              types.String `tfsdk:"id"`
	URL                             types.String `tfsdk:"url"`
	SocketTimeoutMillis             types.Int64  `tfsdk:"socket_timeout_millis"`
	Username                        types.String `tfsdk:"username"`
	Password                        types.String `tfsdk:"password"`
	EnableEventReplication          types.Bool   `tfsdk:"enable_event_replication"`
	Enabled                         types.Bool   `tfsdk:"enabled"`
	CronExp                         types.String `tfsdk:"cron_exp"`
	SyncDeletes                     types.Bool   `tfsdk:"sync_deletes"`
	SyncProperties                  types.Bool   `tfsdk:"sync_properties"`
	SyncStatistics                  types.Bool   `tfsdk:"sync_statistics"`
	RepoKey                         types.String `tfsdk:"repo_key"`
	Proxy                           types.String `tfsdk:"proxy"`
	ReplicationKey                  types.String `tfsdk:"replication_key"`
	IncludePathPrefixPattern        types.String `tfsdk:"include_path_prefix_pattern"`
	ExcludePathPrefixPattern        types.String `tfsdk:"exclude_path_prefix_pattern"`
	CheckBinaryExistenceInFilestore types.Bool   `tfsdk:"check_binary_existence_in_filestore"`
}

func (m LocalRepositorySingleReplicationResourceModel) toAPIModel(_ context.Context, apiModel *LocalSingleReplicationUpdateAPIModel) (diags diag.Diagnostics) {
	*apiModel = LocalSingleReplicationUpdateAPIModel{
		LocalSingleReplicationAPIModel: LocalSingleReplicationAPIModel{
			URL:                             m.URL.ValueString(),
			SocketTimeoutMillis:             m.SocketTimeoutMillis.ValueInt64(),
			Username:                        m.Username.ValueString(),
			Password:                        m.Password.ValueString(),
			EnableEventReplication:          m.EnableEventReplication.ValueBool(),
			Enabled:                         m.Enabled.ValueBool(),
			CronExp:                         m.CronExp.ValueString(),
			SyncDeletes:                     m.SyncDeletes.ValueBool(),
			SyncProperties:                  m.SyncProperties.ValueBool(),
			SyncStatistics:                  m.SyncStatistics.ValueBool(),
			RepoKey:                         m.RepoKey.ValueString(),
			IncludePathPrefixPattern:        m.IncludePathPrefixPattern.ValueString(),
			ExcludePathPrefixPattern:        m.ExcludePathPrefixPattern.ValueString(),
			CheckBinaryExistenceInFilestore: m.CheckBinaryExistenceInFilestore.ValueBool(),
		},
		Proxy: m.Proxy.ValueString(),
	}

	return
}

func (m *LocalRepositorySingleReplicationResourceModel) fromAPIModel(_ context.Context, apiModel LocalSingleReplicationGetAPIModel) (diags diag.Diagnostics) {
	m.ID = types.StringValue(apiModel.RepoKey)
	m.URL = types.StringValue(apiModel.URL)
	m.SocketTimeoutMillis = types.Int64Value(apiModel.SocketTimeoutMillis)
	m.Username = types.StringValue(apiModel.Username)

	if m.Password.IsUnknown() {
		m.Password = types.StringNull()
	}

	m.EnableEventReplication = types.BoolValue(apiModel.EnableEventReplication)
	m.Enabled = types.BoolValue(apiModel.Enabled)
	m.CronExp = types.StringValue(apiModel.CronExp)
	m.SyncDeletes = types.BoolValue(apiModel.SyncDeletes)
	m.SyncProperties = types.BoolValue(apiModel.SyncProperties)
	m.SyncStatistics = types.BoolValue(apiModel.SyncStatistics)
	m.RepoKey = types.StringValue(apiModel.RepoKey)
	m.IncludePathPrefixPattern = types.StringValue(apiModel.IncludePathPrefixPattern)
	m.ExcludePathPrefixPattern = types.StringValue(apiModel.ExcludePathPrefixPattern)
	m.CheckBinaryExistenceInFilestore = types.BoolValue(apiModel.CheckBinaryExistenceInFilestore)
	m.Proxy = types.StringValue(apiModel.ProxyRef)
	m.ReplicationKey = types.StringValue(apiModel.ReplicationKey)

	return
}

type LocalSingleReplicationAPIModel struct {
	URL                             string `json:"url"`
	SocketTimeoutMillis             int64  `json:"socketTimeoutMillis"`
	Username                        string `json:"username"`
	Password                        string `json:"password"`
	EnableEventReplication          bool   `json:"enableEventReplication"`
	Enabled                         bool   `json:"enabled"`
	CronExp                         string `json:"cronExp"`
	SyncDeletes                     bool   `json:"syncDeletes"`
	SyncProperties                  bool   `json:"syncProperties"`
	SyncStatistics                  bool   `json:"syncStatistics"`
	RepoKey                         string `json:"repoKey"`
	IncludePathPrefixPattern        string `json:"includePathPrefixPattern"`
	ExcludePathPrefixPattern        string `json:"excludePathPrefixPattern"`
	CheckBinaryExistenceInFilestore bool   `json:"checkBinaryExistenceInFilestore"`
}

type LocalSingleReplicationGetAPIModel struct {
	LocalSingleReplicationAPIModel
	ProxyRef       string `json:"proxyRef"`
	ReplicationKey string `json:"replicationKey"`
}

type LocalSingleReplicationUpdateAPIModel struct {
	LocalSingleReplicationAPIModel
	Proxy string `json:"proxy"`
}

func (r *LocalRepositorySingleReplicationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *LocalRepositorySingleReplicationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				MarkdownDescription: "A valid CRON expression that you can use to control replication frequency. Eg: `0 0 12 * * ? *`, `0 0 2 ? * MON-SAT *`. Note: use 6 or 7 parts format - Seconds, Minutes Hours, Day Of Month, Month, Day Of Week, Year (optional). Specifying both a day-of-week AND a day-of-month parameter is not supported. One of them should be replaced by `?`. Incorrect: `* 5,7,9 14/2 * * WED,SAT *`, correct: `* 5,7,9 14/2 ? * WED,SAT *`. See details in [Cron Trigger Tutorial](https://www.quartz-scheduler.org/documentation/quartz-2.3.0/tutorials/crontrigger.html).",
			},
			"enable_event_replication": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When set, each event will trigger replication of the artifacts changed in this event. This can be any type of event on artifact, e.g. add, deleted or property change. Default value is `false`.",
			},
			"url": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					validatorfw_string.IsURLHttpOrHttps(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				MarkdownDescription: "The URL of the target local repository on a remote Artifactory server. Use the format `https://<artifactory_url>/artifactory/<repository_name>`.",
			},
			"socket_timeout_millis": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(15000),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				MarkdownDescription: "The network timeout in milliseconds to use for remote operations.",
			},
			"username": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The HTTP authentication username.",
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "Use either the HTTP authentication password or identity token.",
			},
			"sync_deletes": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When set, items that were deleted locally should also be deleted remotely (also applies to properties metadata). Note that enabling this option, will delete artifacts on the target that do not exist in the source repository. Default value is `false`.",
			},
			"sync_properties": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "When set, the task also synchronizes the properties of replicated artifacts. Default value is `true`.",
			},
			"sync_statistics": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When set, the task also synchronizes artifact download statistics. Set to avoid inadvertent cleanup at the target instance when setting up replication for disaster recovery. Default value is `false`.",
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
			"proxy": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "A proxy configuration to use when communicating with the remote instance.",
			},
			"replication_key": schema.StringAttribute{
				Computed:    true,
				Description: "Replication ID. The ID is known only after the replication is created, for this reason it's `Computed` and can not be set by the user in HCL.",
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

func (r *LocalRepositorySingleReplicationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *LocalRepositorySingleReplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan LocalRepositorySingleReplicationResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if verified, err := verifyRepoRclass(plan.RepoKey.ValueString(), "local", r.ProviderData.Client.R()); !verified {
		utilfw.UnableToCreateResourceError(resp, fmt.Sprintf("source repository rclass is not local, only local repositories are supported by this resource %v", err))
		return
	}

	var replication LocalSingleReplicationUpdateAPIModel
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
	if plan.Proxy.IsUnknown() || plan.Proxy.IsNull() {
		plan.Proxy = types.StringValue("") // SDKv2 version uses util.GetString() which returns empty string if attribute not defined/set
	}
	if plan.Password.IsUnknown() || plan.Password.IsNull() {
		plan.Password = types.StringValue("") // SDKv2 version uses util.GetString() which returns empty string if attribute not defined/set
	}
	plan.ReplicationKey = types.StringNull()

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LocalRepositorySingleReplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state LocalRepositorySingleReplicationResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.ProviderData.Client.R().
		SetPathParam("repo_key", state.RepoKey.ValueString()).
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

	var replication LocalSingleReplicationGetAPIModel

	// Artifactory will return an array for Enterprise instances that support multipush replication
	var replications []LocalSingleReplicationGetAPIModel
	err = json.Unmarshal(response.Body(), &replications)
	if err != nil {
		err = json.Unmarshal(response.Body(), &replication)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to marshall result",
				err.Error(),
			)
			return
		}
	} else {
		replication = replications[0]
	}

	resp.Diagnostics.Append(state.fromAPIModel(ctx, replication)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *LocalRepositorySingleReplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan LocalRepositorySingleReplicationResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state LocalRepositorySingleReplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if verified, err := verifyRepoRclass(plan.RepoKey.ValueString(), "local", r.ProviderData.Client.R()); !verified {
		utilfw.UnableToUpdateResourceError(resp, fmt.Sprintf("source repository rclass is not local, only local repositories are supported by this resource %v", err))
		return
	}

	var replication LocalSingleReplicationUpdateAPIModel
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

func (r *LocalRepositorySingleReplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state LocalRepositorySingleReplicationResourceModel

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
func (r *LocalRepositorySingleReplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("repo_key"), req, resp)
}
