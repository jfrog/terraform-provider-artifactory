package security

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const CertificateEndpoint = "artifactory/api/system/security/certificates/"

func NewCertificateResource() resource.Resource {
	return &CertificateResource{}
}

type CertificateResource struct {
	ProviderData utilsdk.ProvderMetadata
}

// CertificateResourceModel describes the Terraform resource data model to match the
// resource schema.
type CertificateResourceModel struct {
	Alias       types.String `tfsdk:"alias"`
	Content     types.String `tfsdk:"content"`
	File        types.String `tfsdk:"file"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	IssuedBy    types.String `tfsdk:"issued_by"`
	IssuedOn    types.String `tfsdk:"issued_on"`
	IssuedTo    types.String `tfsdk:"issued_to"`
	ValidUntil  types.String `tfsdk:"valid_until"`
}

func (r *CertificateResourceModel) FromAPIModel(ctx context.Context, model *CertificateAPIModel) diag.Diagnostics {
	r.Alias = types.StringValue(model.Alias)
	r.Fingerprint = types.StringValue(model.Fingerprint)
	r.IssuedBy = types.StringValue(model.IssuedBy)
	r.IssuedOn = types.StringValue(model.IssuedOn)
	r.IssuedTo = types.StringValue(model.IssuedTo)
	r.ValidUntil = types.StringValue(model.ValidUntil)

	return nil
}

// CertificateAPIModel describes the API data model.
type CertificateAPIModel struct {
	Alias       string `json:"certificateAlias"`
	Fingerprint string `json:"fingerprint"`
	IssuedOn    string `json:"issuedOn"`
	IssuedBy    string `json:"issuedBy"`
	IssuedTo    string `json:"issuedTo"`
	ValidUntil  string `json:"validUntil"`
}

func (r *CertificateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "artifactory_certificate"
}

func (r *CertificateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage the public GPG trusted keys used to verify distributed release bundles.",
		Attributes: map[string]schema.Attribute{
			"alias": schema.StringAttribute{
				MarkdownDescription: "Name of certificate",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "PEM-encoded client certificate and private key",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					contentMustBeX509Certificate(),
				},
			},
			"file": schema.StringAttribute{
				MarkdownDescription: "File system path to PEM file",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					fileMustBeX509Certificate(),
				},
			},
			"fingerprint": schema.StringAttribute{
				MarkdownDescription: "SHA256 fingerprint of the certificate",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"issued_on": schema.StringAttribute{
				MarkdownDescription: "The time & date when the certificate is valid from",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"issued_by": schema.StringAttribute{
				MarkdownDescription: "Name of the certificate authority that issued the certificate",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"issued_to": schema.StringAttribute{
				MarkdownDescription: "Name of whom the certificate has been issued to",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"valid_until": schema.StringAttribute{
				MarkdownDescription: "The time & date when the certificate expires",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r CertificateResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("content"),
			path.MatchRoot("file"),
		),
	}
}

func (r *CertificateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(utilsdk.ProvderMetadata)
}

func updateCertificate(content, file, alias basetypes.StringValue, restyRequest *resty.Request) (*resty.Response, error) {
	// Convert from Terraform data model into API data model
	var contentData string
	if !content.IsNull() {
		contentData = content.ValueString()
	}

	if !file.IsNull() {
		data, err := os.ReadFile(file.ValueString())
		if err != nil {
			return nil, fmt.Errorf("failed to read content from file %s", file.ValueString())
		}

		contentData = string(data)
	}

	response, err := restyRequest.
		SetHeader("content-type", "text/plain").
		SetBody(contentData).
		Post(CertificateEndpoint + alias.ValueString())
	if err != nil {
		return nil, fmt.Errorf("failed to update certificate with Artifactory %v", err.Error())
	}

	return response, nil
}

func (r *CertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *CertificateResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := updateCertificate(plan.Content, plan.File, plan.Alias, r.ProviderData.Client.R())
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode() != http.StatusOK {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	cert, err := FindCertificate(plan.Alias.ValueString(), r.ProviderData.Client.R())
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	resp.Diagnostics.Append(plan.FromAPIModel(ctx, cert)...)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func FindCertificate(alias string, restyRequest *resty.Request) (*CertificateAPIModel, error) {
	var certificates []CertificateAPIModel
	_, err := restyRequest.
		SetResult(&certificates).
		Get(CertificateEndpoint)
	if err != nil {
		return nil, err
	}

	var cert *CertificateAPIModel
	// No way other than to loop through each certificate
	for _, c := range certificates {
		if c.Alias == alias {
			cert = &c
			break
		}
	}

	return cert, nil
}

func (r *CertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *CertificateResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cert, err := FindCertificate(state.Alias.ValueString(), r.ProviderData.Client.R())
	if err != nil {
		resp.Diagnostics.AddError(
			"failed to read certificates from Artifactory",
			err.Error(),
		)
		return
	}

	if cert == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(state.FromAPIModel(ctx, cert)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *CertificateResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := updateCertificate(plan.Content, plan.File, plan.Alias, r.ProviderData.Client.R())
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode() != http.StatusOK {
		utilfw.UnableToUpdateResourceError(resp, response.String())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CertificateResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	response, err := r.ProviderData.Client.R().
		Delete(CertificateEndpoint + state.Alias.ValueString())

	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// Return error if the HTTP status code is not 204 No Content or 404 Not Found
	if response.StatusCode() != http.StatusOK {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *CertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("alias"), req, resp)
}

type certificateContentValidator struct{}

func (v certificateContentValidator) Description(_ context.Context) string {
	return "certificate must be a valid X.509 certificate."
}

func (v certificateContentValidator) MarkdownDescription(_ context.Context) string {
	return "certificate must be a valid X.509 certificate."
}

func (v certificateContentValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the value is unknown or null, there is nothing to validate.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	pemData := req.ConfigValue.ValueString()

	_, err := extractCertificate(pemData)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"invalid certificate format",
			err.Error(),
		)
		return
	}
}

func contentMustBeX509Certificate() certificateContentValidator {
	return certificateContentValidator{}
}

type certificateFileValidator struct{}

func (v certificateFileValidator) Description(_ context.Context) string {
	return "file must be a valid X.509 certificate."
}

func (v certificateFileValidator) MarkdownDescription(_ context.Context) string {
	return "file must be a valid X.509 certificate."
}

func (v certificateFileValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the value is unknown or null, there is nothing to validate.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	filePath := req.ConfigValue.ValueString()

	if _, err := os.Stat(filePath); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"file does not exist",
			err.Error(),
		)
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"failed to read file",
			err.Error(),
		)
		return
	}

	_, err = extractCertificate(string(data))
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"invalid certificate format",
			err.Error(),
		)
		return
	}
}

func fileMustBeX509Certificate() certificateFileValidator {
	return certificateFileValidator{}
}

func extractCertificate(pemData string) (*x509.Certificate, error) {
	block, rest := pem.Decode([]byte(pemData))
	for block != nil {
		if block.Type != "CERTIFICATE" {
			block, rest = pem.Decode(rest)
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}

		return cert, nil
	}

	return nil, fmt.Errorf("no certificate in PEM data")
}
