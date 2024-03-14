package configuration

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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

	"gopkg.in/yaml.v3"
)

type ProxyAPIModel struct {
	Key               string `xml:"key" yaml:"-"`
	Host              string `xml:"host" yaml:"host"`
	Port              int64  `xml:"port" yaml:"port"`
	Username          string `xml:"username" yaml:"username"`
	Password          string `xml:"password" yaml:"password"`
	NtHost            string `xml:"ntHost" yaml:"ntHost"`
	NtDomain          string `xml:"domain" yaml:"domain"`
	PlatformDefault   bool   `xml:"platformDefault" yaml:"platformDefault"`
	RedirectedToHosts string `xml:"redirectedToHosts" yaml:"redirectedToHosts"`
	Services          string `xml:"services" yaml:"services"`
}

func (p ProxyAPIModel) Id() string {
	return p.Key
}

type ProxiesAPIModel struct {
	Proxies []ProxyAPIModel `xml:"proxies>proxy" yaml:"proxy"`
}

type ProxyResourceModel struct {
	Key               types.String `tfsdk:"key"`
	Host              types.String `tfsdk:"host"`
	Port              types.Int64  `tfsdk:"port"`
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
	NtHost            types.String `tfsdk:"nt_host"`
	NtDomain          types.String `tfsdk:"nt_domain"`
	PlatformDefault   types.Bool   `tfsdk:"platform_default"`
	RedirectedToHosts types.Set    `tfsdk:"redirect_to_hosts"`
	Services          types.Set    `tfsdk:"services"`
}

func (r *ProxyResourceModel) toAPIModel(ctx context.Context, proxy *ProxyAPIModel) diag.Diagnostics {
	var redirectedToHosts []string
	diags := r.RedirectedToHosts.ElementsAs(ctx, &redirectedToHosts, true)
	if diags != nil {
		return diags
	}

	var services []string
	diags = r.Services.ElementsAs(ctx, &services, true)
	if diags != nil {
		return diags
	}

	*proxy = ProxyAPIModel{
		Key:               r.Key.ValueString(),
		Host:              r.Host.ValueString(),
		Port:              r.Port.ValueInt64(),
		Username:          r.Username.ValueString(),
		Password:          r.Password.ValueString(),
		NtHost:            r.NtHost.ValueString(),
		NtDomain:          r.NtDomain.ValueString(),
		PlatformDefault:   r.PlatformDefault.ValueBool(),
		RedirectedToHosts: strings.Join(redirectedToHosts, ","),
		Services:          strings.Join(services, ","),
	}

	return nil
}

func (r *ProxyResourceModel) FromAPIModel(ctx context.Context, proxy *ProxyAPIModel) diag.Diagnostics {
	r.Key = types.StringValue(proxy.Key)
	r.Host = types.StringValue(proxy.Host)
	r.Port = types.Int64Value(proxy.Port)
	r.Username = types.StringValue(proxy.Username)
	r.NtHost = types.StringValue(proxy.NtHost)
	r.NtDomain = types.StringValue(proxy.NtDomain)
	r.PlatformDefault = types.BoolValue(proxy.PlatformDefault)

	if proxy.RedirectedToHosts != "" {
		redirectedToHosts, diags := types.SetValueFrom(ctx, types.StringType, strings.Split(proxy.RedirectedToHosts, ","))
		if diags != nil {
			return diags
		}
		r.RedirectedToHosts = redirectedToHosts
	}

	if proxy.Services != "" {
		services, diags := types.SetValueFrom(ctx, types.StringType, strings.Split(proxy.Services, ","))
		if diags != nil {
			return diags
		}
		r.Services = services
	}

	return nil
}

func NewProxyResource() resource.Resource {
	return &ProxyResource{
		TypeName: "artifactory_proxy",
	}
}

type ProxyResource struct {
	ProviderData util.ProvderMetadata
	TypeName     string
}

func (r *ProxyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *ProxyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides an Artifactory Proxy resource. This resource configuration is only available for self-hosted instance. It corresponds to 'proxies' config block in system configuration XML (REST endpoint: artifactory/api/system/configuration).",
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The unique ID of the proxy.",
			},
			"host": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "The name of the proxy host.",
			},
			"port": schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.Between(0, 65535),
				},
				MarkdownDescription: "The proxy port number.",
			},
			"username": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The proxy username when authentication credentials are required.",
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The proxy password when authentication credentials are required.",
			},
			"nt_host": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The computer name of the machine (the machine connecting to the NTLM proxy).",
			},
			"nt_domain": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The proxy domain/realm name.",
			},
			"platform_default": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "When set, this proxy will be the default proxy for new remote repositories and for internal HTTP requests issued by Artifactory. Will also be used as proxy for all other services in the platform (for example: Xray, Distribution, etc).",
			},
			"redirect_to_hosts": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "An optional list of host names to which this proxy may redirect requests. The credentials defined for the proxy are reused by requests redirected to all of these hosts.",
			},
			"services": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.OneOf([]string{"jfrt", "jfmc", "jfxr", "jfds"}...)),
				},
				Description: "An optional list of services names to which this proxy be the default of. The options are jfrt, jfmc, jfxr, jfds",
			},
		},
	}
}

func (r *ProxyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProvderMetadata)
}

func (r ProxyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data ProxyResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If platform_default is not configured, return without warning.
	if data.PlatformDefault.IsNull() || data.PlatformDefault.IsUnknown() {
		return
	}

	// If services is not configured, return without warning.
	if data.Services.IsNull() || data.Services.IsUnknown() {
		return
	}

	if data.PlatformDefault.ValueBool() && len(data.Services.Elements()) > 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("services"),
			"Invalid Attribute Configuration",
			"services cannot be set when platform_default is true",
		)
	}
}

func (r *ProxyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client, r.ProviderData.ProductId, r.TypeName)

	var plan *ProxyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var proxy ProxyAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &proxy)...)
	if resp.Diagnostics.HasError() {
		return
	}

	///* EXPLANATION FOR BELOW CONSTRUCTION USAGE.
	//There is a difference in xml structure usage between GET and PATCH calls of API: /artifactory/api/system/configuration.
	//GET call structure has "propertySets -> propertySet -> Array of property sets".
	//PATCH call structure has "propertySets -> propertySet (dynamic sting). Property name and predefinedValues names are also dynamic strings".
	//Following nested map of string structs are constructed to match the usage of PATCH call with the consideration of dynamic strings.
	//*/
	var body = map[string]map[string]ProxyAPIModel{
		"proxies": {
			proxy.Key: proxy,
		},
	}

	content, err := yaml.Marshal(&body)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	err = SendConfigurationPatch(content, r.ProviderData)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProxyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client, r.ProviderData.ProductId, r.TypeName)

	var state *ProxyResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var proxies ProxiesAPIModel
	response, err := r.ProviderData.Client.R().
		SetResult(&proxies).
		Get("artifactory/api/system/configuration")
	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, "failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	matchedProxyConfig := FindConfigurationById[ProxyAPIModel](proxies.Proxies, state.Key.ValueString())
	if matchedProxyConfig == nil {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("key"),
			"no matching proxy found",
			state.Key.ValueString(),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(state.FromAPIModel(ctx, matchedProxyConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ProxyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client, r.ProviderData.ProductId, r.TypeName)

	var plan *ProxyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var proxy ProxyAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &proxy)...)
	if resp.Diagnostics.HasError() {
		return
	}

	///* EXPLANATION FOR BELOW CONSTRUCTION USAGE.
	//There is a difference in xml structure usage between GET and PATCH calls of API: /artifactory/api/system/configuration.
	//GET call structure has "propertySets -> propertySet -> Array of property sets".
	//PATCH call structure has "propertySets -> propertySet (dynamic sting). Property name and predefinedValues names are also dynamic strings".
	//Following nested map of string structs are constructed to match the usage of PATCH call with the consideration of dynamic strings.
	//*/
	var body = map[string]map[string]ProxyAPIModel{
		"proxies": {
			proxy.Key: proxy,
		},
	}

	content, err := yaml.Marshal(&body)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	err = SendConfigurationPatch(content, r.ProviderData)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProxyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client, r.ProviderData.ProductId, r.TypeName)

	var data ProxyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteBackupConfig := fmt.Sprintf(`
proxies:
  %s: ~
`, data.Key.ValueString())

	err := SendConfigurationPatch([]byte(deleteBackupConfig), r.ProviderData)
	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *ProxyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}
