package security

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const DistributionPublicKeysAPIEndPoint = "artifactory/api/security/keys/trusted"

func NewDistributionPublicKeyResource() resource.Resource {
	return &DistributionPublicKeyResource{}
}

type DistributionPublicKeyResource struct {
	ProviderData utilsdk.ProvderMetadata
}

// DistributionPublicKeyResourceModel describes the Terraform resource data model to match the
// resource schema.
type DistributionPublicKeyResourceModel struct {
	KeyId       types.String          `tfsdk:"key_id"`
	Alias       types.String          `tfsdk:"alias"`
	Fingerprint types.String          `tfsdk:"fingerprint"`
	PublicKey   TablessPublicKeyValue `tfsdk:"public_key"`
	IssuedOn    types.String          `tfsdk:"issued_on"`
	IssuedBy    types.String          `tfsdk:"issued_by"`
	ValidUntil  types.String          `tfsdk:"valid_until"`
}

func (r *DistributionPublicKeyResourceModel) FromAPIModel(ctx context.Context, model *DistributionPublicKeyAPIModel) diag.Diagnostics {
	r.KeyId = types.StringValue(model.KeyId)
	r.Alias = types.StringValue(model.Alias)
	r.Fingerprint = types.StringValue(model.Fingerprint)
	r.PublicKey = tablessPublicKeyValue(model.PublicKey)
	r.IssuedOn = types.StringValue(model.IssuedOn)
	r.IssuedBy = types.StringValue(model.IssuedBy)
	r.ValidUntil = types.StringValue(model.ValidUntil)

	return nil
}

// DistributionPublicKeyAPIModel describes the API data model.
type DistributionPublicKeyAPIModel struct {
	KeyId       string `json:"kid,omitempty"`
	Alias       string `json:"alias"`
	Fingerprint string `json:"fingerprint,omitempty"`
	PublicKey   string `json:"key"`
	IssuedOn    string `json:"issued_on,omitempty"`
	IssuedBy    string `json:"issued_by,omitempty"`
	ValidUntil  string `json:"valid_until,omitempty"`
}

type DistributionPublicKeysList struct {
	Keys []DistributionPublicKeyAPIModel `json:"keys"`
}

func (r *DistributionPublicKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "artifactory_distribution_public_key"
}

func (r *DistributionPublicKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage the public GPG trusted keys used to verify distributed release bundles.",
		Attributes: map[string]schema.Attribute{
			"key_id": schema.StringAttribute{
				MarkdownDescription: "Returns the key id by which this key is referenced in Artifactory.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: "Will be used as an identifier when uploading/retrieving the public key via REST API.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"fingerprint": schema.StringAttribute{
				MarkdownDescription: "Returns the computed key fingerprint",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "The Public key to add as a trusted distribution GPG key. To avoid state drift, ensure there are no leading tab or space characters for each line.",
				Required:            true,
				CustomType:          TablessPublicKeyType{},
				Validators:          []validator.String{publicKeyMustBeGPGOrRSA()},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"issued_on": schema.StringAttribute{
				MarkdownDescription: "Returns the date/time when this GPG key was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"issued_by": schema.StringAttribute{
				MarkdownDescription: "Returns the name and eMail address of issuer.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"valid_until": schema.StringAttribute{
				MarkdownDescription: "Returns the date/time when this GPG key expires.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *DistributionPublicKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(utilsdk.ProvderMetadata)
}

func (r *DistributionPublicKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *DistributionPublicKeyResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	var publicKey DistributionPublicKeyAPIModel

	body := DistributionPublicKeyAPIModel{
		Alias:     plan.Alias.ValueString(),
		PublicKey: stripTabs(plan.PublicKey.ValueString()),
	}

	response, err := r.ProviderData.Client.R().
		SetBody(body).
		SetResult(&publicKey).
		Post(DistributionPublicKeysAPIEndPoint)

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode() != http.StatusCreated {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	// Assign the resource ID for the resource in the state
	resp.Diagnostics.Append(plan.FromAPIModel(ctx, &publicKey)...)
	// data.KeyId = types.StringValue(publicKey.KeyId)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DistributionPublicKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *DistributionPublicKeyResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	var publicKeys DistributionPublicKeysList

	response, err := r.ProviderData.Client.R().
		SetResult(&publicKeys).
		Get(DistributionPublicKeysAPIEndPoint)

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if err != nil {
		if response.StatusCode() == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
		}
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	for _, key := range publicKeys.Keys {
		if key.Alias == state.Alias.ValueString() {
			resp.Diagnostics.Append(state.FromAPIModel(ctx, &key)...)
			tflog.Debug(ctx, fmt.Sprintf("state after: %v", state))
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DistributionPublicKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update is not supported
}

func (r *DistributionPublicKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DistributionPublicKeyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	response, err := r.ProviderData.Client.R().
		Delete(fmt.Sprintf("%s/%s", DistributionPublicKeysAPIEndPoint, state.KeyId.ValueString()))

	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// Return error if the HTTP status code is not 204 No Content or 404 Not Found
	if response.StatusCode() != http.StatusNoContent && response.StatusCode() != http.StatusOK {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *DistributionPublicKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("alias"), req, resp)
}

type publicKeyValidator struct{}

func (v publicKeyValidator) Description(_ context.Context) string {
	return "public key must be either PGP or RSA."
}

func (v publicKeyValidator) MarkdownDescription(_ context.Context) string {
	return "public key must be either PGP or RSA."
}

func (v publicKeyValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the value is unknown or null, there is nothing to validate.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	stripped := strings.ReplaceAll(req.ConfigValue.ValueString(), "\t", "")
	// currently can't validate GPG
	if strings.Contains(stripped, "BEGIN PGP PUBLIC KEY BLOCK") {
		resp.Diagnostics.AddAttributeWarning(
			req.Path,
			"Usage of GPG can't be validated.",
			"Due to limitations of go libraries, your GPG key can't be validated client side.",
		)
		return
	}

	pubPem, _ := pem.Decode([]byte(stripped))
	if pubPem == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"RSA public key not in pem format.",
			"RSA public key not in pem format.",
		)
		return
	}

	if !strings.Contains(pubPem.Type, "PUBLIC KEY") {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"RSA public key is of the wrong type.",
			fmt.Sprintf("RSA public keymust container the header 'PUBLIC KEY': Pem Type: %s ", pubPem.Type),
		)
		return
	}

	var parsedKey interface{}
	if _, err := x509.ParsePKIXPublicKey(pubPem.Bytes); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Unable to parse RSA public key.",
			err.Error(),
		)
		return
	}

	if _, ok := parsedKey.(*rsa.PublicKey); !ok {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Unable to cast to RSA public key data type.",
			"",
		)
		return
	}
}

func publicKeyMustBeGPGOrRSA() publicKeyValidator {
	return publicKeyValidator{}
}

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.StringTypable = TablessPublicKeyType{}

type TablessPublicKeyType struct {
	basetypes.StringType
}

func (t TablessPublicKeyType) Equal(o attr.Type) bool {
	other, ok := o.(TablessPublicKeyType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t TablessPublicKeyType) String() string {
	return "TablessPublicKeyType"
}

func (t TablessPublicKeyType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	// TablessPublicKeyValue defined in the value type section
	value := TablessPublicKeyValue{
		StringValue: in,
	}

	return value, nil
}

func (t TablessPublicKeyType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t TablessPublicKeyType) ValueType(ctx context.Context) attr.Value {
	// CustomStringValue defined in the value type section
	return TablessPublicKeyValue{}
}

func tablessPublicKeyValue(value string) TablessPublicKeyValue {
	return TablessPublicKeyValue{
		StringValue: basetypes.NewStringValue(value),
	}
}

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.StringValuableWithSemanticEquals = TablessPublicKeyValue{}

type TablessPublicKeyValue struct {
	basetypes.StringValue
}

func (v TablessPublicKeyValue) Equal(o attr.Value) bool {
	other, ok := o.(TablessPublicKeyValue)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v TablessPublicKeyValue) Type(ctx context.Context) attr.Type {
	// CustomStringType defined in the schema type section
	return TablessPublicKeyType{}
}

// StringSemanticEquals returns true if the given string value is semantically equal to the current string value. (case-insensitive)
func (v TablessPublicKeyValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(TablessPublicKeyValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}

	tflog.Debug(ctx, fmt.Sprintf("newValue.ValueString(): %s", newValue.ValueString()))
	tflog.Debug(ctx, fmt.Sprintf("v.ValueString(): %s", v.ValueString()))

	return strings.EqualFold(stripTabs(newValue.ValueString()), v.ValueString()), diags
}
