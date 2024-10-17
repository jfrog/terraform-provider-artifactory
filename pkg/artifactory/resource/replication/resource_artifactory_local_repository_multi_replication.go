package replication

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	"github.com/samber/lo"

	"github.com/jfrog/terraform-provider-shared/client"
)

var _ resource.Resource = &LocalRepositoryMultiReplicationResource{}

func NewLocalRepositoryMultiReplicationResource() resource.Resource {
	return &LocalRepositoryMultiReplicationResource{
		TypeName: "artifactory_local_repository_multi_replication",
	}
}

type LocalRepositoryMultiReplicationResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

type LocalRepositoryMultiReplicationResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	RepoKey                types.String `tfsdk:"repo_key"`
	EnableEventReplication types.Bool   `tfsdk:"enable_event_replication"`
	CronExp                types.String `tfsdk:"cron_exp"`
	Replication            types.List   `tfsdk:"replication"`
}

func (m LocalRepositoryMultiReplicationResourceModel) toAPIModel(_ context.Context, apiModel *LocalMultiReplicationUpdateAPIModel) (diags diag.Diagnostics) {
	replications := lo.Map(
		m.Replication.Elements(),
		func(elem attr.Value, _ int) ReplicationUpdateAPIModel {
			attrs := elem.(types.Object).Attributes()

			return ReplicationUpdateAPIModel{
				ReplicationAPIModel: ReplicationAPIModel{
					URL:                             attrs["url"].(types.String).ValueString(),
					SocketTimeoutMillis:             attrs["socket_timeout_millis"].(types.Int64).ValueInt64(),
					Username:                        attrs["username"].(types.String).ValueString(),
					Password:                        attrs["password"].(types.String).ValueString(),
					CronExp:                         m.CronExp.ValueString(),
					RepoKey:                         m.RepoKey.ValueString(),
					Enabled:                         attrs["enabled"].(types.Bool).ValueBool(),
					SyncDeletes:                     attrs["sync_deletes"].(types.Bool).ValueBool(),
					SyncProperties:                  attrs["sync_properties"].(types.Bool).ValueBool(),
					SyncStatistics:                  attrs["sync_statistics"].(types.Bool).ValueBool(),
					IncludePathPrefixPattern:        attrs["include_path_prefix_pattern"].(types.String).ValueString(),
					ExcludePathPrefixPattern:        attrs["exclude_path_prefix_pattern"].(types.String).ValueString(),
					CheckBinaryExistenceInFilestore: attrs["check_binary_existence_in_filestore"].(types.Bool).ValueBool(),
				},
				Proxy:        attrs["proxy"].(types.String).ValueString(),
				DisableProxy: attrs["disable_proxy"].(types.Bool).ValueBool(),
			}
		},
	)

	*apiModel = LocalMultiReplicationUpdateAPIModel{
		CronExp:                m.CronExp.ValueString(),
		EnableEventReplication: m.EnableEventReplication.ValueBool(),
		Replications:           replications,
	}

	return
}

var replicationResourceModelAttributeTypes map[string]attr.Type = map[string]attr.Type{
	"url":                                 types.StringType,
	"socket_timeout_millis":               types.Int64Type,
	"username":                            types.StringType,
	"password":                            types.StringType,
	"enabled":                             types.BoolType,
	"sync_deletes":                        types.BoolType,
	"sync_properties":                     types.BoolType,
	"sync_statistics":                     types.BoolType,
	"include_path_prefix_pattern":         types.StringType,
	"exclude_path_prefix_pattern":         types.StringType,
	"check_binary_existence_in_filestore": types.BoolType,
	"proxy":                               types.StringType,
	"disable_proxy":                       types.BoolType,
	"replication_key":                     types.StringType,
}

var replicationListResourceModelAttributeTypes types.ObjectType = types.ObjectType{
	AttrTypes: replicationResourceModelAttributeTypes,
}

func (m *LocalRepositoryMultiReplicationResourceModel) fromAPIModel(_ context.Context, apiModel []ReplicationGetAPIModel) (diags diag.Diagnostics) {
	replications := lo.Map(
		apiModel,
		func(replication ReplicationGetAPIModel, idx int) attr.Value {
			if idx == 0 {
				m.ID = types.StringValue(replication.RepoKey)
				m.RepoKey = types.StringValue(replication.RepoKey)
				m.CronExp = types.StringValue(replication.CronExp)
				m.EnableEventReplication = types.BoolValue(replication.EnableEventReplication)
			}

			// set password from current state to avoid state drift
			// from missing password in Artifactory API response
			password := types.StringNull()
			stateReplication, _, found := lo.FindIndexOf(
				m.Replication.Elements(),
				func(elem attr.Value) bool {
					attrs := elem.(types.Object).Attributes()
					return attrs["url"].Equal(types.StringValue(replication.URL))
				},
			)
			if found {
				password = stateReplication.(types.Object).Attributes()["password"].(types.String)
			}

			r, ds := types.ObjectValue(
				replicationResourceModelAttributeTypes,
				map[string]attr.Value{
					"url":                                 types.StringValue(replication.URL),
					"socket_timeout_millis":               types.Int64Value(replication.SocketTimeoutMillis),
					"username":                            types.StringValue(replication.Username),
					"password":                            password,
					"enabled":                             types.BoolValue(replication.Enabled),
					"sync_deletes":                        types.BoolValue(replication.SyncDeletes),
					"sync_properties":                     types.BoolValue(replication.SyncProperties),
					"sync_statistics":                     types.BoolValue(replication.SyncStatistics),
					"include_path_prefix_pattern":         types.StringValue(replication.IncludePathPrefixPattern),
					"exclude_path_prefix_pattern":         types.StringValue(replication.ExcludePathPrefixPattern),
					"check_binary_existence_in_filestore": types.BoolValue(replication.CheckBinaryExistenceInFilestore),
					"proxy":                               types.StringValue(replication.ProxyRef),
					"disable_proxy":                       types.BoolValue(replication.DisableProxy),
					"replication_key":                     types.StringValue(replication.ReplicationKey),
				},
			)

			if ds != nil {
				diags = append(diags, ds...)
			}

			return r
		},
	)

	r, ds := types.ListValue(
		replicationListResourceModelAttributeTypes,
		replications,
	)
	if ds != nil {
		diags = append(diags, ds...)
	}

	m.Replication = r

	return
}

type ReplicationAPIModel struct {
	Username                        string `json:"username"`
	Password                        string `json:"password"`
	URL                             string `json:"url"`
	CronExp                         string `json:"cronExp"`
	RepoKey                         string `json:"repoKey"`
	EnableEventReplication          bool   `json:"enableEventReplication"`
	SocketTimeoutMillis             int64  `json:"socketTimeoutMillis"`
	Enabled                         bool   `json:"enabled"`
	SyncDeletes                     bool   `json:"syncDeletes"`
	SyncProperties                  bool   `json:"syncProperties"`
	SyncStatistics                  bool   `json:"syncStatistics"`
	IncludePathPrefixPattern        string `json:"includePathPrefixPattern"`
	ExcludePathPrefixPattern        string `json:"excludePathPrefixPattern"`
	CheckBinaryExistenceInFilestore bool   `json:"checkBinaryExistenceInFilestore"`
}

type ReplicationGetAPIModel struct {
	ReplicationAPIModel
	ProxyRef       string `json:"proxyRef"`
	ReplicationKey string `json:"replicationKey"`
	DisableProxy   bool   `json:"disableProxy"`
}

type ReplicationUpdateAPIModel struct {
	ReplicationAPIModel
	Proxy        string `json:"proxy"`
	DisableProxy bool   `json:"disableProxy"`
}

type LocalMultiReplicationUpdateAPIModel struct {
	CronExp                string                      `json:"cronExp,omitempty"`
	EnableEventReplication bool                        `json:"enableEventReplication"`
	Replications           []ReplicationUpdateAPIModel `json:"replications,omitempty"`
}

func (r *LocalRepositoryMultiReplicationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *LocalRepositoryMultiReplicationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
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
		},
		Blocks: map[string]schema.Block{
			"replication": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
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
							Computed:  true,
							Default:   stringdefault.StaticString(""),
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
							Description: "Use either the HTTP authentication password or identity token (https://www.jfrog.com/confluence/display/JFROG/User+Profile#UserProfile-IdentityTokenidentitytoken).",
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
						"disable_proxy": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							MarkdownDescription: "When set to `true`, the `proxy` attribute will be ignored (from version 7.41.7). The default value is `false`.",
						},
						"replication_key": schema.StringAttribute{
							Computed:    true,
							Description: "Replication ID. The ID is known only after the replication is created, for this reason it's `Computed` and can not be set by the user in HCL.",
						},
						"check_binary_existence_in_filestore": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
							MarkdownDescription: "Enabling the `check_binary_existence_in_filestore` flag requires an Enterprise+ license. When true, enables distributed checksum storage. For more information, see " +
								"[Optimizing Repository Replication with Checksum-Based Storage](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-OptimizingRepositoryReplicationUsingStorageLevelSynchronizationOptions).",
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
		MarkdownDescription: "Provides a local repository replication resource, also referred to as Artifactory push replication. This can be used to create and manage Artifactory local repository replications using [Multi-push Replication API](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-CreateorReplaceLocalMulti-pushReplication).\n\n" +
			"Push replication is used to synchronize Local Repositories, and is implemented by the Artifactory server on the near end invoking a synchronization of artifacts to the far end.\n\n" +
			"See the [Official Documentation](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-PushReplication).\n\n" +
			"This resource replaces `artifactory_push_replication` and used to create a replication of one local repository to multiple repositories on the remote server.\n\n" +
			"~> This resource requires Artifactory Enterprise license. Use `artifactory_local_repository_single_replication` with other licenses.",
	}
}

func (r *LocalRepositoryMultiReplicationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *LocalRepositoryMultiReplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan LocalRepositoryMultiReplicationResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if verified, err := verifyRepoRclass(plan.RepoKey.ValueString(), "local", r.ProviderData.Client.R()); !verified {
		utilfw.UnableToCreateResourceError(resp, fmt.Sprintf("source repository rclass is not local, only local repositories are supported by this resource %v", err))
		return
	}

	var replication LocalMultiReplicationUpdateAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &replication)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.ProviderData.Client.R().
		SetPathParam("repo_key", plan.RepoKey.ValueString()).
		SetBody(replication).
		Put(MultiReplicationEndpoint)

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	// Read the replication data back for 'replication_key'
	var replications []ReplicationGetAPIModel

	response, err = r.ProviderData.Client.R().
		SetPathParam("repo_key", plan.RepoKey.ValueString()).
		SetResult(&replications).
		Get(ReplicationEndpoint)

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	resp.Diagnostics.Append(plan.fromAPIModel(ctx, replications)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LocalRepositoryMultiReplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state LocalRepositoryMultiReplicationResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var replications []ReplicationGetAPIModel

	response, err := r.ProviderData.Client.R().
		SetPathParam("repo_key", state.RepoKey.ValueString()).
		SetResult(&replications).
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

	resp.Diagnostics.Append(state.fromAPIModel(ctx, replications)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *LocalRepositoryMultiReplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan LocalRepositoryMultiReplicationResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state LocalRepositoryMultiReplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if verified, err := verifyRepoRclass(plan.RepoKey.ValueString(), "local", r.ProviderData.Client.R()); !verified {
		utilfw.UnableToUpdateResourceError(resp, fmt.Sprintf("source repository rclass is not local, only local repositories are supported by this resource %v", err))
		return
	}

	var replication LocalMultiReplicationUpdateAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &replication)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.ProviderData.Client.R().
		SetPathParam("repo_key", plan.RepoKey.ValueString()).
		SetBody(replication).
		AddRetryCondition(client.RetryOnMergeError).
		Post(MultiReplicationEndpoint)

	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToUpdateResourceError(resp, response.String())
		return
	}

	// Read the replication data back for 'replication_key'
	var replications []ReplicationGetAPIModel

	response, err = r.ProviderData.Client.R().
		SetPathParam("repo_key", plan.RepoKey.ValueString()).
		SetResult(&replications).
		Get(ReplicationEndpoint)

	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToUpdateResourceError(resp, response.String())
		return
	}

	resp.Diagnostics.Append(plan.fromAPIModel(ctx, replications)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LocalRepositoryMultiReplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state LocalRepositoryMultiReplicationResourceModel

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
func (r *LocalRepositoryMultiReplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("repo_key"), req, resp)
}
