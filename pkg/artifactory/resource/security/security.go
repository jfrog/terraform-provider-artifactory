package security

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func VerifyKeyPair(id string, request *resty.Request) (*resty.Response, error) {
	return request.Head(KeypairEndPoint + id)
}

func stripTabs(val string) string {
	return strings.ReplaceAll(val, "\t", "")
}

type signingKeyValidator struct{}

func (v signingKeyValidator) Description(_ context.Context) string {
	return "public key must be either PGP or RSA."
}

func (v signingKeyValidator) MarkdownDescription(_ context.Context) string {
	return "public key must be either PGP or RSA."
}

func (v signingKeyValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the value is unknown or null, there is nothing to validate.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	stripped := stripTabs(req.ConfigValue.ValueString())
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
	parsedKey, err := x509.ParsePKIXPublicKey(pubPem.Bytes)
	if err != nil {
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

func signingKeyMustBeGPGOrRSA() signingKeyValidator {
	return signingKeyValidator{}
}

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.StringTypable = TablessSigningKeyType{}

type TablessSigningKeyType struct {
	basetypes.StringType
}

func (t TablessSigningKeyType) Equal(o attr.Type) bool {
	other, ok := o.(TablessSigningKeyType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t TablessSigningKeyType) String() string {
	return "TablessSigningKeyType"
}

func (t TablessSigningKeyType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	// TablessPublicKeyValue defined in the value type section
	value := TablessSigningKeyValue{
		StringValue: in,
	}

	return value, nil
}

func (t TablessSigningKeyType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t TablessSigningKeyType) ValueType(ctx context.Context) attr.Value {
	// CustomStringValue defined in the value type section
	return TablessSigningKeyValue{}
}

func tablessSigningKeyValue(value string) TablessSigningKeyValue {
	return TablessSigningKeyValue{
		StringValue: basetypes.NewStringValue(value),
	}
}

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.StringValuableWithSemanticEquals = TablessSigningKeyValue{}

type TablessSigningKeyValue struct {
	basetypes.StringValue
}

func (v TablessSigningKeyValue) Equal(o attr.Value) bool {
	other, ok := o.(TablessSigningKeyValue)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v TablessSigningKeyValue) Type(ctx context.Context) attr.Type {
	// CustomStringType defined in the schema type section
	return TablessSigningKeyType{}
}

// StringSemanticEquals returns true if the given string value is semantically equal to the current string value. (case-insensitive)
func (v TablessSigningKeyValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(TablessSigningKeyValue)
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

	return strings.EqualFold(stripTabs(newValue.ValueString()), v.ValueString()), diags
}
