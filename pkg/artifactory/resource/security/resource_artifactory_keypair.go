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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const KeypairEndPoint = "artifactory/api/security/keypair/"

func NewKeyPairResource() resource.Resource {
	return &KeyPairResource{}
}

type KeyPairResource struct {
	ProviderData utilsdk.ProvderMetadata
}

// KeyPairResourceModel describes the Terraform resource data model to match the
// resource schema.
type KeyPairResourceModel struct {
	PairName   types.String           `tfsdk:"pair_name"`
	PairType   types.String           `tfsdk:"pair_type"`
	Alias      types.String           `tfsdk:"alias"`
	PrivateKey TablessSigningKeyValue `tfsdk:"private_key"`
	Passphrase types.String           `tfsdk:"passphrase"`
	PublicKey  TablessSigningKeyValue `tfsdk:"public_key"`
}

func (r *KeyPairResourceModel) FromAPIModel(ctx context.Context, model *KeyPairAPIModel) diag.Diagnostics {
	r.PairName = types.StringValue(model.PairName)
	r.PairType = types.StringValue(model.PairType)
	r.Alias = types.StringValue(model.Alias)
	r.PublicKey = tablessSigningKeyValue(model.PublicKey)

	return nil
}

// KeyPairAPIModel describes the API data model.
type KeyPairAPIModel struct {
	PairName   string `json:"pairName"`
	PairType   string `json:"pairType"`
	Alias      string `json:"alias"`
	PrivateKey string `json:"privateKey"`
	Passphrase string `json:"passphrase"`
	PublicKey  string `json:"publicKey"`
}

func (r *KeyPairResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "artifactory_keypair"
}

func (r *KeyPairResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "RSA key pairs are used to sign and verify the Alpine Linux index files in JFrog Artifactory, " +
			"while GPG key pairs are used to sign and validate packages integrity in JFrog Distribution. " +
			"The JFrog Platform enables you to manage multiple RSA and GPG signing keys through the Keys Management UI " +
			"and REST API. The JFrog Platform supports managing multiple pairs of GPG signing keys to sign packages for" +
			" authentication of several package types such as Debian, Opkg, and RPM through the Keys Management UI and REST API.",
		Attributes: map[string]schema.Attribute{
			"pair_name": schema.StringAttribute{
				MarkdownDescription: "A unique identifier for the Key Pair record.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pair_type": schema.StringAttribute{
				MarkdownDescription: "Key Pair type. Supported types - GPG and RSA.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"RSA", "GPG"}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: "Will be used as a filename when retrieving the public key via REST API",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"private_key": schema.StringAttribute{
				MarkdownDescription: "Private key. PEM format will be validated. Must not include extranous spaces or tabs.",
				Required:            true,
				Sensitive:           true,
				CustomType:          TablessSigningKeyType{},
				Validators: []validator.String{
					privateKeyMustValid(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"passphrase": schema.StringAttribute{
				MarkdownDescription: "Passphrase will be used to decrypt the private key. Validated server side.",
				Optional:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "Public key. PEM format will be validated. Must not include extranous spaces or tabs.",
				Required:            true,
				CustomType:          TablessSigningKeyType{},
				Validators: []validator.String{
					signingKeyMustBeGPGOrRSA(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *KeyPairResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(utilsdk.ProvderMetadata)
}

func (r *KeyPairResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *KeyPairResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyPair := KeyPairAPIModel{
		PairName:   plan.PairName.ValueString(),
		PairType:   plan.PairType.ValueString(),
		Alias:      plan.Alias.ValueString(),
		PrivateKey: plan.PrivateKey.ValueString(),
		Passphrase: plan.Passphrase.ValueString(),
		PublicKey:  plan.PublicKey.ValueString(),
	}

	response, err := r.ProviderData.Client.R().
		SetBody(keyPair).
		Post(KeypairEndPoint)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	// Return error if the HTTP status code is not 201 Created
	if response.StatusCode() != http.StatusCreated {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *KeyPairResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state KeyPairResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	var keyPair KeyPairAPIModel

	response, err := r.ProviderData.Client.R().
		SetResult(&keyPair).
		Get(KeypairEndPoint + state.PairName.ValueString())

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if err != nil {
		if response.StatusCode() == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(state.FromAPIModel(ctx, &keyPair)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *KeyPairResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// noop
}

func (r *KeyPairResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state KeyPairResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	response, err := r.ProviderData.Client.R().
		Delete(KeypairEndPoint + state.PairName.ValueString())

	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// Return error if the HTTP status code is not 200 OK or 404 Not Found
	if response.StatusCode() != http.StatusNotFound && response.StatusCode() != http.StatusOK {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *KeyPairResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("pair_name"), req, resp)
}

type privateKeyValidator struct{}

func (v privateKeyValidator) Description(_ context.Context) string {
	return "private key must be valid."
}

func (v privateKeyValidator) MarkdownDescription(_ context.Context) string {
	return "private key must be valid."
}

func (v privateKeyValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the value is unknown or null, there is nothing to validate.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	stripped := stripTabs(req.ConfigValue.ValueString())
	// currently can't validate GPG
	if strings.Contains(stripped, "BEGIN PGP PRIVATE KEY BLOCK") {
		resp.Diagnostics.AddAttributeWarning(
			req.Path,
			"Usage of GPG can't be validated.",
			"Due to limitations of go libraries, your GPG key can't be validated client side.",
		)
		return
	}

	privatePem, _ := pem.Decode([]byte(stripped))
	if privatePem == nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"unable to decode private key pem format",
			"",
		)
		return
	}

	if privatePem.Type != "RSA PRIVATE KEY" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"RSA private key is of the wrong type.",
			fmt.Sprintf("Pem Type: %s", privatePem.Type),
		)
		return
	}

	var parsedKey interface{}
	parsedKey, err := x509.ParsePKCS1PrivateKey(privatePem.Bytes)
	if err != nil {
		parsedKey, err = x509.ParsePKCS8PrivateKey(privatePem.Bytes)
		if err != nil {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Unable to parse RSA private key.",
				err.Error(),
			)
			return
		}
	}

	if _, ok := parsedKey.(*rsa.PrivateKey); !ok {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"unable to cast to RSA private key",
			"",
		)
		return
	}
}

func privateKeyMustValid() privateKeyValidator {
	return privateKeyValidator{}
}
