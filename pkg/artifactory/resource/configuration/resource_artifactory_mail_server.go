package configuration

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"gopkg.in/yaml.v3"
)

type MailServerAPIModel struct {
	Enabled        bool   `xml:"enabled" yaml:"enabled"`
	ArtifactoryURL string `xml:"artifactoryUrl" yaml:"artifactoryUrl"`
	From           string `xml:"from" yaml:"from"`
	Host           string `xml:"host" yaml:"host"`
	Username       string `xml:"username" yaml:"username"`
	Password       string `xml:"password" yaml:"password"`
	Port           int64  `xml:"port" yaml:"port"`
	SubjectPrefix  string `xml:"subjectPrefix" yaml:"subjectPrefix"`
	UseSSL         bool   `xml:"ssl" yaml:"ssl"`
	UseTLS         bool   `xml:"tls" yaml:"tls"`
}

type MailServer struct {
	Server *MailServerAPIModel `xml:"mailServer"`
}

type MailServerResourceModel struct {
	Enabled        types.Bool   `tfsdk:"enabled"`
	ArtifactoryURL types.String `tfsdk:"artifactory_url"`
	From           types.String `tfsdk:"from"`
	Host           types.String `tfsdk:"host"`
	Username       types.String `tfsdk:"username"`
	Password       types.String `tfsdk:"password"`
	Port           types.Int64  `tfsdk:"port"`
	SubjectPrefix  types.String `tfsdk:"subject_prefix"`
	UseSSL         types.Bool   `tfsdk:"use_ssl"`
	UseTLS         types.Bool   `tfsdk:"use_tls"`
}

func (r *MailServerResourceModel) ToAPIModel(ctx context.Context, mailServer *MailServerAPIModel) diag.Diagnostics {
	// Convert from Terraform resource model into API model
	*mailServer = MailServerAPIModel{
		Enabled:        r.Enabled.ValueBool(),
		ArtifactoryURL: r.ArtifactoryURL.ValueString(),
		From:           r.From.ValueString(),
		Host:           r.Host.ValueString(),
		Username:       r.Username.ValueString(),
		Password:       r.Password.ValueString(),
		Port:           r.Port.ValueInt64(),
		SubjectPrefix:  r.SubjectPrefix.ValueString(),
		UseSSL:         r.UseSSL.ValueBool(),
		UseTLS:         r.UseTLS.ValueBool(),
	}

	return nil
}

func (r *MailServerResourceModel) FromAPIModel(ctx context.Context, mailServer *MailServerAPIModel) diag.Diagnostics {
	r.Enabled = types.BoolValue(mailServer.Enabled)
	r.ArtifactoryURL = types.StringValue(mailServer.ArtifactoryURL)
	r.From = types.StringValue(mailServer.From)
	r.Host = types.StringValue(mailServer.Host)
	r.Username = types.StringValue(mailServer.Username)
	r.Port = types.Int64Value(mailServer.Port)
	r.SubjectPrefix = types.StringValue(mailServer.SubjectPrefix)
	r.UseSSL = types.BoolValue(mailServer.UseSSL)
	r.UseTLS = types.BoolValue(mailServer.UseTLS)

	return nil
}

func NewMailServerResource() resource.Resource {
	return &MailServerResource{}
}

type MailServerResource struct {
	ProviderData utilsdk.ProvderMetadata
}

func (r *MailServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "artifactory_mail_server"
}

func (r *MailServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provides an Artifactory mail server config resource. This resource configuration corresponds to mail server config block in system configuration XML (REST endpoint: artifactory/api/system/configuration). Manages mail server settings of the Artifactory instance.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "When set, mail notifications are enabled.",
				Required:            true,
			},
			"artifactory_url": schema.StringAttribute{
				MarkdownDescription: "The Artifactory URL to to link to in all outgoing messages.",
				Optional:            true,
				Validators: []validator.String{
					validatorfw_string.IsURLHttpOrHttps(),
				},
			},
			"from": schema.StringAttribute{
				MarkdownDescription: "The 'from' address header to use in all outgoing messages.",
				Optional:            true,
				Validators: []validator.String{
					validatorfw_string.IsEmail(),
				},
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "The mail server IP address / DNS.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The username for authentication with the mail server.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password for authentication with the mail server.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "The port number of the mail server.",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.AtMost(65535),
				},
			},
			"subject_prefix": schema.StringAttribute{
				MarkdownDescription: "A prefix to use for the subject of all outgoing mails.",
				Optional:            true,
			},
			"use_ssl": schema.BoolAttribute{
				MarkdownDescription: "When set to 'true', uses a secure connection to the mail server.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"use_tls": schema.BoolAttribute{
				MarkdownDescription: "When set to 'true', uses Transport Layer Security when connecting to the mail server.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *MailServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(utilsdk.ProvderMetadata)
}

func (r *MailServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *MailServerResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var mailServer MailServerAPIModel
	resp.Diagnostics.Append(plan.ToAPIModel(ctx, &mailServer)...)
	if resp.Diagnostics.HasError() {
		return
	}

	/* EXPLANATION FOR BELOW CONSTRUCTION USAGE.

	There is a difference in xml structure usage between GET and PATCH calls of API: /artifactory/api/system/configuration.

	GET call structure has "backups -> backup -> Array of backup config blocks".

	PATCH call structure has "backups -> Name/Key of backup that is being patched -> config block of the backup being patched".

	Since the Name/Key is dynamic string, following nested map of string structs are constructed to match the usage of PATCH call.

	See https://www.jfrog.com/confluence/display/JFROG/Artifactory+YAML+Configuration for patching system configuration
	using YAML
	*/
	var constructBody = map[string]MailServerAPIModel{}
	constructBody["mailServer"] = mailServer
	content, err := yaml.Marshal(&constructBody)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	err = SendConfigurationPatch(content, r.ProviderData)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	// Assign the resource ID for the resource in the state
	plan.Host = types.StringValue(mailServer.Host)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MailServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *MailServerResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var mailServer MailServer
	_, err := r.ProviderData.Client.R().
		SetResult(&mailServer).
		Get("artifactory/api/system/configuration")
	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, "failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		return
	}

	if mailServer.Server == nil {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("host"),
			"no mail server found",
			"",
		)
		resp.State.RemoveResource(ctx)
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(state.FromAPIModel(ctx, mailServer.Server)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *MailServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *MailServerResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	// Convert from Terraform data model into API data model
	var mailServer MailServerAPIModel
	resp.Diagnostics.Append(plan.ToAPIModel(ctx, &mailServer)...)

	/* EXPLANATION FOR BELOW CONSTRUCTION USAGE.

	There is a difference in xml structure usage between GET and PATCH calls of API: /artifactory/api/system/configuration.

	GET call structure has "backups -> backup -> Array of backup config blocks".

	PATCH call structure has "backups -> Name/Key of backup that is being patched -> config block of the backup being patched".

	Since the Name/Key is dynamic string, following nested map of string structs are constructed to match the usage of PATCH call.

	See https://www.jfrog.com/confluence/display/JFROG/Artifactory+YAML+Configuration for patching system configuration
	using YAML
	*/
	var constructBody = map[string]MailServerAPIModel{}
	constructBody["mailServer"] = mailServer
	content, err := yaml.Marshal(&constructBody)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	err = SendConfigurationPatch(content, r.ProviderData)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	resp.Diagnostics.Append(plan.FromAPIModel(ctx, &mailServer)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MailServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MailServerResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteMailServerConfig := `mailServer: ~`

	err := SendConfigurationPatch([]byte(deleteMailServerConfig), r.ProviderData)
	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *MailServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// "host" attribute is used here but it's a noop. There's only ever one mail server on Artifactory
	// so there's no need to use ID to fetch.
	resource.ImportStatePassthroughID(ctx, path.Root("host"), req, resp)
}
