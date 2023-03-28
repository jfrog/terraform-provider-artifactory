package security

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/jfrog/terraform-provider-shared/util"
)

const KeypairEndPoint = "artifactory/api/security/keypair/"

type KeyPairPayLoad struct {
	PairName    string `hcl:"pair_name" json:"pairName"`
	PairType    string `hcl:"pair_type" json:"pairType"`
	Alias       string `hcl:"alias" json:"alias"`
	PrivateKey  string `hcl:"private_key" json:"privateKey"`
	Passphrase  string `hcl:"passphrase" json:"passphrase"`
	PublicKey   string `hcl:"public_key" json:"publicKey"`
	Unavailable bool   `hcl:"unavailable" json:"unavailable"`
}

var keyPairSchema = map[string]*schema.Schema{
	"pair_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "A unique identifier for the Key Pair record.",
	},
	"pair_type": {
		Type: schema.TypeString,
		// working sample PGP key is checked in but not tested
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"RSA", "GPG"}, false)),
		Required:         true,
		Description:      "Key Pair type. Supported types - GPG and RSA.",
		ForceNew:         true,
	},
	"alias": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Will be used as a filename when retrieving the public key via REST API",
		ForceNew:    true,
	},
	"private_key": {
		Type:             schema.TypeString,
		Sensitive:        true,
		Required:         true,
		StateFunc:        stripTabs,
		ValidateDiagFunc: validatePrivateKey,
		Description:      "Private key. PEM format will be validated.",
		ForceNew:         true,
	},
	"passphrase": {
		Type:             schema.TypeString,
		Optional:         true,
		ForceNew:         true,
		DiffSuppressFunc: ignoreEmpty,
		Sensitive:        true,
		Description:      "Passphrase will be used to decrypt the private key. Validated server side",
	},
	"public_key": {
		Type:             schema.TypeString,
		Required:         true,
		StateFunc:        stripTabs,
		ValidateDiagFunc: validatePublicKey,
		ForceNew:         true,
		Description:      "Public key. PEM format will be validated.",
	},
	"unavailable": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Unknown usage. Returned in the json payload and cannot be set.",
	},
}

func ResourceArtifactoryKeyPair() *schema.Resource {
	return &schema.Resource{
		CreateContext: createKeyPair,
		ReadContext:   readKeyPair,
		DeleteContext: rmKeyPair,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "RSA key pairs are used to sign and verify the Alpine Linux index files in JFrog Artifactory, " +
			"while GPG key pairs are used to sign and validate packages integrity in JFrog Distribution. " +
			"The JFrog Platform enables you to manage multiple RSA and GPG signing keys through the Keys Management UI " +
			"and REST API. The JFrog Platform supports managing multiple pairs of GPG signing keys to sign packages for" +
			" authentication of several package types such as Debian, Opkg, and RPM through the Keys Management UI and REST API.",

		Schema: keyPairSchema,
	}
}

func validatePrivateKey(value interface{}, _ cty.Path) diag.Diagnostics {
	stripped := strings.ReplaceAll(value.(string), "\t", "")
	var err error
	// currently can't validate GPG
	if strings.Contains(stripped, "BEGIN PGP PRIVATE KEY BLOCK") {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "usage of GPG can't be validated.",
				Detail:   "Due to limitations of go libraries, your GPG key can't be validated client side",
			},
		}
	}
	privPem, _ := pem.Decode([]byte(stripped))
	if privPem == nil {
		return diag.Errorf("unable to decode private key pem format")
	}
	var privPemBytes []byte
	if privPem.Type != "RSA PRIVATE KEY" {
		return diag.Errorf("RSA private key is of the wrong type. Pem Type: %s", privPem.Type)
	}

	privPemBytes = privPem.Bytes
	var parsedKey interface{}
	if parsedKey, err = x509.ParsePKCS1PrivateKey(privPemBytes); err != nil {
		if parsedKey, err = x509.ParsePKCS8PrivateKey(privPemBytes); err != nil { // note this returns type `interface{}`
			return diag.FromErr(err)
		}
	}

	_, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return diag.Errorf("unable to cast to RSA private key")
	}
	return nil
}

func validatePublicKey(value interface{}, _ cty.Path) diag.Diagnostics {
	var err error

	stripped := strings.ReplaceAll(value.(string), "\t", "")
	// currently can't validate GPG
	if strings.Contains(stripped, "BEGIN PGP PUBLIC KEY BLOCK") {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "usage of GPG can't be validated.",
				Detail:   "Due to limitations of go libraries, your GPG key can't be validated client side",
			},
		}
	}
	pubPem, _ := pem.Decode([]byte(stripped))
	if pubPem == nil {
		return diag.Errorf("rsa public key not in pem format")
	}
	if !strings.Contains(pubPem.Type, "PUBLIC KEY") {
		return diag.Errorf("RSA public key is of the wrong type and must container the header 'PUBLIC KEY': Pem Type: %s ", pubPem.Type)
	}
	var parsedKey interface{}

	if parsedKey, err = x509.ParsePKIXPublicKey(pubPem.Bytes); err != nil {
		return diag.Errorf("unable to parse RSA public key")
	}

	if _, ok := parsedKey.(*rsa.PublicKey); !ok {
		return diag.Errorf("unable to cast to RSA public key")
	}

	return nil
}

func stripTabs(val interface{}) string {
	return strings.ReplaceAll(val.(string), "\t", "")
}

func ignoreEmpty(_, _, _ string, _ *schema.ResourceData) bool {
	return false
}

func unpackKeyPair(s *schema.ResourceData) (interface{}, string, error) {
	d := &util.ResourceData{ResourceData: s}
	result := KeyPairPayLoad{
		PairName:    d.GetString("pair_name", false),
		PairType:    d.GetString("pair_type", false),
		Alias:       d.GetString("alias", false),
		Passphrase:  d.GetString("passphrase", false),
		PrivateKey:  strings.ReplaceAll(d.GetString("private_key", false), "\t", ""),
		PublicKey:   strings.ReplaceAll(d.GetString("public_key", false), "\t", ""),
		Unavailable: d.GetBool("unavailable", false),
	}
	return &result, result.PairName, nil
}

var keyPairPacker = packer.Universal(
	predicate.All(
		predicate.Ignore("private_key", "passphrase"),
		predicate.SchemaHasKey(keyPairSchema),
	),
)

func createKeyPair(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	keyPair, key, _ := unpackKeyPair(d)

	_, err := m.(util.ProvderMetadata).Client.R().
		AddRetryCondition(client.RetryOnMergeError).
		SetBody(keyPair).
		Post(KeypairEndPoint)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(key)
	return readKeyPair(ctx, d, m)
}

func readKeyPair(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	data := KeyPairPayLoad{}
	resp, err := meta.(util.ProvderMetadata).Client.R().SetResult(&data).Get(KeypairEndPoint + d.Id())
	if err != nil {
		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	err = keyPairPacker(data, d)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func rmKeyPair(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	_, err := m.(util.ProvderMetadata).Client.R().Delete(KeypairEndPoint + d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func (kp KeyPairPayLoad) Id() string {
	return kp.PairName
}
