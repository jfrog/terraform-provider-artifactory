package artifact

import (
	"context"
	"encoding/base64"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfpath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
)

func NewArtifactResource() resource.Resource {
	return &ArtifactResource{
		TypeName: "artifactory_artifact",
	}
}

type ArtifactResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

type ArtifactResourceModel struct {
	Repository     types.String `tfsdk:"repository"`
	Path           types.String `tfsdk:"path"`
	FilePath       types.String `tfsdk:"file_path"`
	ContentBase64  types.String `tfsdk:"content_base64"`
	ChecksumMD5    types.String `tfsdk:"checksum_md5"`
	ChecksumSHA1   types.String `tfsdk:"checksum_sha1"`
	ChecksumSHA256 types.String `tfsdk:"checksum_sha256"`
	Created        types.String `tfsdk:"created"`
	CreatedBy      types.String `tfsdk:"created_by"`
	DownloadURI    types.String `tfsdk:"download_uri"`
	MimeType       types.String `tfsdk:"mime_type"`
	Size           types.Int64  `tfsdk:"size"`
	URI            types.String `tfsdk:"uri"`
}

func (r *ArtifactResourceModel) LocalFilePath() (string, error) {
	if !r.FilePath.IsNull() && !r.FilePath.IsUnknown() {
		return r.FilePath.ValueString(), nil
	}

	f, err := os.CreateTemp("", "artifactory_artifact_")
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(r.ContentBase64.ValueString())
	if err != nil {
		return "", err
	}

	if _, err = f.Write(data); err != nil {
		return "", err
	}

	if err = f.Sync(); err != nil {
		return "", err
	}

	return f.Name(), nil
}

func (r *ArtifactResourceModel) fromAPIModel(apiModel ArtifactAPIModel) diag.Diagnostics {
	r.Repository = types.StringValue(apiModel.Repository)
	r.Path = types.StringValue(apiModel.Path)
	r.ChecksumMD5 = types.StringValue(apiModel.Checksums.MD5)
	r.ChecksumSHA1 = types.StringValue(apiModel.Checksums.SHA1)
	r.ChecksumSHA256 = types.StringValue(apiModel.Checksums.SHA256)
	r.Created = types.StringValue(apiModel.Created)
	r.CreatedBy = types.StringValue(apiModel.CreatedBy)
	r.DownloadURI = types.StringValue(apiModel.DownloadURI)
	r.MimeType = types.StringValue(apiModel.MimeType)

	size, err := strconv.ParseInt(apiModel.Size, 10, 64)
	if err != nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"failed to convert size to Int",
				err.Error(),
			),
		}
	}
	r.Size = types.Int64Value(size)
	r.URI = types.StringValue(apiModel.URI)

	return nil
}

type ArtifactChecksumsAPIModel struct {
	MD5    string `json:"md5"`
	SHA1   string `json:"sha1"`
	SHA256 string `json:"sha256"`
}

type ArtifactAPIModel struct {
	Repository  string                    `json:"repo"`
	Path        string                    `json:"path"`
	Checksums   ArtifactChecksumsAPIModel `json:"checksums"`
	Created     string                    `json:"created"`
	CreatedBy   string                    `json:"createdBy"`
	DownloadURI string                    `json:"downloadUri"`
	MimeType    string                    `json:"mimeType"`
	Size        string                    `json:"size"`
	URI         string                    `json:"uri"`
}

func (r *ArtifactResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *ArtifactResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"repository": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				MarkdownDescription: "Name of the respository.",
			},
			"path": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.RegexMatches(regexp.MustCompile(`^\/.+$`), "Path must start with '/'"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				MarkdownDescription: "The relative path in the target repository. Must begin with a '/'. You can add key-value matrix parameters to deploy the artifacts with properties. For more details, please refer to [Introducing Matrix Parameters](https://jfrog.com/help/r/jfrog-artifactory-documentation/using-properties-in-deployment-and-resolution).",
			},
			"file_path": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(tfpath.MatchRoot("content_base64")),
					stringvalidator.LengthAtLeast(1),
					fileExistValidator{},
				},
				MarkdownDescription: "Path to the source file. Conflicts with `content_base64`. Either one of these attribute must be set.",
			},
			"content_base64": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(tfpath.MatchRoot("file_path")),
					stringvalidator.LengthAtLeast(1),
				},
				MarkdownDescription: "Base64 content of the source file. Conflicts with `file_path`. Either one of these attribute must be set.",
			},
			"checksum_md5": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "MD5 checksum of the artifact.",
			},
			"checksum_sha1": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "SHA1 checksum of the artifact.",
			},
			"checksum_sha256": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "SHA256 checksum of the artifact.",
			},
			"created": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when artifact is created.",
			},
			"created_by": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "User who deploys the artifact.",
			},
			"download_uri": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Download URI of the artifact.",
			},
			"mime_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "MIME type of the artifact.",
			},
			"size": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Size of the artifact, in bytes.",
			},
			"uri": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "URI of the artifact.",
			},
		},
		MarkdownDescription: "Provides a resource for deploying artifact to Artifactory repository. Support deploying a single artifact only. Changes to `repository` or `path` attributes will trigger a recreation of the resource (i.e. delete then create). See [JFrog documentation](https://jfrog.com/help/r/jfrog-artifactory-documentation/deploy-a-single-artifact) for more details.",
	}
}

func (r *ArtifactResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *ArtifactResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ArtifactResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result ArtifactAPIModel

	// upload file to Artifactory repo
	repo_target_path := path.Join(plan.Repository.ValueString(), plan.Path.ValueString())
	localFilePath, err := plan.LocalFilePath()
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}
	if !plan.ContentBase64.IsNull() {
		defer os.Remove(localFilePath)
	}

	// open the file as stream
	f, err := os.Open(localFilePath)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}
	defer f.Close()

	response, err := r.ProviderData.Client.R().
		SetRawPathParam("repo_target_path", repo_target_path).
		SetHeader("Content-Type", "application/octet-stream").
		SetBody(f).
		SetResult(&result).
		Put("/artifactory/{repo_target_path}")

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	resp.Diagnostics.Append(plan.fromAPIModel(result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ArtifactResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ArtifactResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	var artifact ArtifactAPIModel
	repo_path := path.Join(state.Repository.ValueString(), state.Path.ValueString())
	response, err := r.ProviderData.Client.R().
		SetRawPathParam("repo_path", repo_path).
		SetResult(&artifact).
		Get("/artifactory/api/storage/{repo_path}")

	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if response.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(state.fromAPIModel(artifact)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ArtifactResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ArtifactResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result ArtifactAPIModel

	// upload file to Artifactory repo
	repo_target_path := path.Join(plan.Repository.ValueString(), plan.Path.ValueString())
	localFilePath, err := plan.LocalFilePath()
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}
	if !plan.ContentBase64.IsNull() {
		defer os.Remove(localFilePath)
	}

	// open the file as stream
	f, err := os.Open(localFilePath)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}
	defer f.Close()

	response, err := r.ProviderData.Client.R().
		SetRawPathParam("repo_target_path", repo_target_path).
		SetHeader("Content-Type", "application/octet-stream").
		SetBody(f).
		SetResult(&result).
		Put("/artifactory/{repo_target_path}")

	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToUpdateResourceError(resp, response.String())
		return
	}

	resp.Diagnostics.Append(plan.fromAPIModel(result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ArtifactResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ArtifactResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	repo_path := path.Join(state.Repository.ValueString(), state.Path.ValueString())
	response, err := r.ProviderData.Client.R().
		SetRawPathParam("repo_path", repo_path).
		Delete("/artifactory/{repo_path}")

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

type fileExistValidator struct{}

func (v fileExistValidator) Description(ctx context.Context) string {
	return "file path must refer to an existing file"
}

func (v fileExistValidator) MarkdownDescription(ctx context.Context) string {
	return "file path must refer to an existing file"
}

func (v fileExistValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the value is unknown or null, there is nothing to validate.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	if _, err := os.Stat(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid file path",
			err.Error(),
		)
	}
}
